package appintel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
	"github.com/gorilla/websocket"
)

var (
	hrefRegex       = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'#\s]+)["']`)
	formActionRegex = regexp.MustCompile(`(?i)<form\b[^>]*\saction\s*=\s*["']([^"']*)["']`)
	inputNameRegex  = regexp.MustCompile(`(?i)<input\b[^>]*\sname\s*=\s*["']([^"']+)["']`)
)

// AutoExplorerConfig specifies parameters for autonomous site interaction and crawling.
type AutoExplorerConfig struct {
	Target       *models.Target
	TrafficRepo  *database.TrafficRepo
	ScanID       string
	StartURL     string
	MaxPages     int
	MaxDepth     int
	PageTimeout  time.Duration
	TotalTimeout time.Duration
	PreferChrome bool
}

// ExplorerProgress tracks active discovery and interaction counts.
type ExplorerProgress struct {
	TargetID        string   `json:"target_id"`
	CurrentURL      string   `json:"current_url"`
	PagesVisited    int      `json:"pages_visited"`
	ElementsClicked int      `json:"elements_clicked"`
	TrafficCaptured int      `json:"traffic_captured"`
	CategoriesFound []string `json:"categories_found"`
	IsComplete      bool     `json:"is_complete"`
	EngineUsed      string   `json:"engine_used"` // "chromium_cdp" or "http_spider"
	Error           string   `json:"error,omitempty"`
}

// AutoExplorer autonomously navigates a web target, clicks buttons, fills forms,
// and records all underlying network transactions into the application intelligence layer.
type AutoExplorer struct {
	config   AutoExplorerConfig
	visited  map[string]bool
	mu       sync.Mutex
	progress ExplorerProgress
	onUpdate func(ExplorerProgress)
}

// NewAutoExplorer creates a new autonomous explorer instance.
func NewAutoExplorer(cfg AutoExplorerConfig, onUpdate func(ExplorerProgress)) *AutoExplorer {
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 25
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 3
	}
	if cfg.PageTimeout <= 0 {
		cfg.PageTimeout = 10 * time.Second
	}
	if cfg.TotalTimeout <= 0 {
		cfg.TotalTimeout = 3 * time.Minute
	}
	if cfg.StartURL == "" && cfg.Target != nil {
		cfg.StartURL = "https://" + cfg.Target.Domain
	}

	return &AutoExplorer{
		config:   cfg,
		visited:  make(map[string]bool),
		onUpdate: onUpdate,
		progress: ExplorerProgress{
			TargetID:        cfg.Target.ID,
			CurrentURL:      cfg.StartURL,
			CategoriesFound: make([]string, 0),
		},
	}
}

// GetProgress returns the current exploration state snapshot.
func (e *AutoExplorer) GetProgress() ExplorerProgress {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.progress
}

// Run executes the autonomous discovery and interaction pipeline.
func (e *AutoExplorer) Run(ctx context.Context) (*ExplorerProgress, error) {
	totalCtx, cancel := context.WithTimeout(ctx, e.config.TotalTimeout)
	defer cancel()

	// Step 1: Check if Chromium is available and requested
	chromePath, chromeErr := FindChromium()
	canUseChrome := e.config.PreferChrome && chromeErr == nil && chromePath != ""

	if canUseChrome {
		e.progress.EngineUsed = "chromium_cdp"
		err := e.runChromiumExplorer(totalCtx, chromePath)
		if err == nil {
			e.finalize(nil)
			return &e.progress, nil
		}
		// Fallback to pure-Go HTTP Spider on CDP failure
	}

	e.progress.EngineUsed = "http_spider"
	err := e.runHTTPSpider(totalCtx)
	e.finalize(err)
	return &e.progress, err
}

func (e *AutoExplorer) finalize(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.progress.IsComplete = true
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		e.progress.Error = err.Error()
	}

	// Calculate captured traffic stats
	if e.config.TrafficRepo != nil && e.config.Target != nil {
		stats, _ := e.config.TrafficRepo.GetCategoryStats(e.config.Target.ID)
		if stats != nil {
			cats := make([]string, 0)
			for c := range stats {
				cats = append(cats, c)
			}
			e.progress.CategoriesFound = cats
		}
		_, total, _ := e.config.TrafficRepo.ListByTarget(e.config.Target.ID, "", "", "", 1, 0)
		e.progress.TrafficCaptured = total
	}

	if e.onUpdate != nil {
		e.onUpdate(e.progress)
	}
}

// runHTTPSpider performs automated crawling and form submission using Go's recording transport.
func (e *AutoExplorer) runHTTPSpider(ctx context.Context) error {
	cfg := InterceptorConfig{
		TargetID:     e.config.Target.ID,
		TargetDomain: e.config.Target.Domain,
		ScanID:       e.config.ScanID,
		TrafficRepo:  e.config.TrafficRepo,
	}

	recordingTransport := NewRecordingTransport(cfg, &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        20,
	})

	client := &http.Client{
		Transport: recordingTransport,
		Timeout:   e.config.PageTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			// Maintain scope on redirects
			if !tools.IsInScope(req.URL.Host, e.config.Target.Domain) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	type queueItem struct {
		url   string
		depth int
	}

	queue := []queueItem{{url: e.config.StartURL, depth: 1}}
	e.visited[normalizeURL(e.config.StartURL)] = true

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if e.progress.PagesVisited >= e.config.MaxPages {
			break
		}

		item := queue[0]
		queue = queue[1:]

		e.mu.Lock()
		e.progress.CurrentURL = item.url
		e.progress.PagesVisited++
		if e.onUpdate != nil {
			e.onUpdate(e.progress)
		}
		e.mu.Unlock()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) BHF-AutoExplorer/2.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
		_ = resp.Body.Close()
		bodyStr := string(bodyBytes)

		// 1. Extract and follow in-scope links
		if item.depth < e.config.MaxDepth {
			matches := hrefRegex.FindAllStringSubmatch(bodyStr, -1)
			for _, m := range matches {
				if len(m) < 2 {
					continue
				}
				resolved := resolveURL(item.url, m[1])
				if resolved == "" {
					continue
				}
				parsed, err := url.Parse(resolved)
				if err != nil || !tools.IsInScope(parsed.Host, e.config.Target.Domain) {
					continue
				}
				norm := normalizeURL(resolved)
				if !e.visited[norm] {
					e.visited[norm] = true
					queue = append(queue, queueItem{url: resolved, depth: item.depth + 1})
				}
			}
		}

		// 2. Discover and interact with forms (submit benign probe inputs)
		formMatches := formActionRegex.FindAllStringSubmatch(bodyStr, -1)
		for _, fm := range formMatches {
			if len(fm) < 2 {
				continue
			}
			actionURL := resolveURL(item.url, fm[1])
			if actionURL == "" {
				actionURL = item.url
			}

			parsedAction, err := url.Parse(actionURL)
			if err != nil || !tools.IsInScope(parsedAction.Host, e.config.Target.Domain) {
				continue
			}

			// Extract input fields and build probe form submission
			inputs := inputNameRegex.FindAllStringSubmatch(bodyStr, -1)
			formValues := url.Values{}
			for _, in := range inputs {
				if len(in) >= 2 {
					formValues.Set(in[1], "bhf-audit-test")
				}
			}

			if len(formValues) > 0 {
				formReq, err := http.NewRequestWithContext(ctx, http.MethodPost, actionURL, strings.NewReader(formValues.Encode()))
				if err == nil {
					formReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					formReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) BHF-AutoExplorer/2.0")
					formResp, err := client.Do(formReq)
					if err == nil {
						_, _ = io.Copy(io.Discard, formResp.Body)
						_ = formResp.Body.Close()
						e.mu.Lock()
						e.progress.ElementsClicked++
						e.mu.Unlock()
					}
				}
			}
		}
	}

	return nil
}

// runChromiumExplorer launches headless Chromium with proxy interception,
// loads the page DOM, evaluates JavaScript, and clicks elements autonomously.
func (e *AutoExplorer) runChromiumExplorer(ctx context.Context, binaryPath string) error {
	// 1. Start dedicated ProxyServer on local dynamic port
	proxyCfg := InterceptorConfig{
		TargetID:     e.config.Target.ID,
		TargetDomain: e.config.Target.Domain,
		ScanID:       e.config.ScanID,
		TrafficRepo:  e.config.TrafficRepo,
	}
	proxy := NewProxyServer(proxyCfg)
	proxyAddr, err := proxy.Start("127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start explorer proxy: %w", err)
	}
	defer func() { _ = proxy.Stop() }()

	// 2. Allocate free port for Chrome DevTools Protocol
	cdpPort := getFreeTCPPort()

	// 3. Launch isolated headless Chromium pointing to our proxy
	browserCfg := BrowserConfig{
		BinaryPath:         binaryPath,
		Headless:           true,
		ProxyAddr:          proxyAddr,
		RemoteDebuggingPort: cdpPort,
		Timeout:            e.config.PageTimeout,
	}

	controller := NewBrowserController(browserCfg)
	session, err := controller.Launch(ctx, "about:blank")
	if err != nil {
		return fmt.Errorf("failed to launch chromium explorer: %w", err)
	}
	defer func() { _ = session.Close() }()

	// 4. Connect to DevTools WebSocket
	wsURL, err := getCDPWebSocketURL(ctx, cdpPort)
	if err != nil {
		return fmt.Errorf("failed to obtain CDP websocket: %w", err)
	}

	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial CDP websocket: %w", err)
	}
	defer wsConn.Close()

	// Enable Page & Runtime domains
	_ = sendCDPCommand(wsConn, 1, "Page.enable", nil)
	_ = sendCDPCommand(wsConn, 2, "Runtime.enable", nil)

	type queueItem struct {
		url   string
		depth int
	}

	queue := []queueItem{{url: e.config.StartURL, depth: 1}}
	e.visited[normalizeURL(e.config.StartURL)] = true

	msgID := 10

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if e.progress.PagesVisited >= e.config.MaxPages {
			break
		}

		item := queue[0]
		queue = queue[1:]

		e.mu.Lock()
		e.progress.CurrentURL = item.url
		e.progress.PagesVisited++
		if e.onUpdate != nil {
			e.onUpdate(e.progress)
		}
		e.mu.Unlock()

		// Navigate to page
		msgID++
		navParams := map[string]any{"url": item.url}
		_ = sendCDPCommand(wsConn, msgID, "Page.navigate", navParams)

		// Wait for initial render and network activity
		time.Sleep(1500 * time.Millisecond)

		// Evaluate exploration script inside the live DOM:
		// - Fills form fields
		// - Clicks buttons, tabs, interactive elements
		// - Scrolls down to trigger AJAX/Fetch calls
		// - Extracts all links
		interactionScript := `(() => {
			let links = [];
			try {
				links = Array.from(document.querySelectorAll('a[href]'))
					.map(a => a.href)
					.filter(h => h.startsWith('http'));
			} catch(e) {}

			// Fill inputs with test values
			try {
				document.querySelectorAll('input').forEach(inp => {
					let t = (inp.type || 'text').toLowerCase();
					if (['text', 'search', 'email'].includes(t) && !inp.value) {
						inp.value = 'bhf-audit-test';
					}
				});
			} catch(e) {}

			// Click visible interactive buttons and tabs
			let clicked = 0;
			try {
				let btns = Array.from(document.querySelectorAll('button, [role="button"], .nav-link, .tab'));
				btns.slice(0, 8).forEach(btn => {
					try {
						if (btn.offsetParent !== null) {
							btn.click();
							clicked++;
						}
					} catch(e) {}
				});
			} catch(e) {}

			// Scroll page
			try {
				window.scrollTo(0, document.body.scrollHeight / 2);
				window.scrollTo(0, document.body.scrollHeight);
			} catch(e) {}

			return { links: links, clicked: clicked };
		})()`

		msgID++
		evalParams := map[string]any{
			"expression":    interactionScript,
			"returnByValue": true,
		}
		res, err := callCDPEval(wsConn, msgID, evalParams)
		if err == nil && res != nil {
			e.mu.Lock()
			e.progress.ElementsClicked += res.Clicked
			e.mu.Unlock()

			if item.depth < e.config.MaxDepth {
				for _, link := range res.Links {
					parsed, err := url.Parse(link)
					if err != nil || !tools.IsInScope(parsed.Host, e.config.Target.Domain) {
						continue
					}
					norm := normalizeURL(link)
					if !e.visited[norm] {
						e.visited[norm] = true
						queue = append(queue, queueItem{url: link, depth: item.depth + 1})
					}
				}
			}
		}

		// Wait briefly after interaction for XHR responses
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// Helpers

type cdpEvalResult struct {
	Links   []string `json:"links"`
	Clicked int      `json:"clicked"`
}

func callCDPEval(conn *websocket.Conn, id int, params map[string]any) (*cdpEvalResult, error) {
	_ = sendCDPCommand(conn, id, "Runtime.evaluate", params)

	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		var resp struct {
			ID     int `json:"id"`
			Result struct {
				Result struct {
					Value cdpEvalResult `json:"value"`
				} `json:"result"`
			} `json:"result"`
		}

		if err := json.Unmarshal(message, &resp); err == nil && resp.ID == id {
			return &resp.Result.Result.Value, nil
		}
	}
}

func sendCDPCommand(conn *websocket.Conn, id int, method string, params map[string]any) error {
	payload := map[string]any{
		"id":     id,
		"method": method,
	}
	if params != nil {
		payload["params"] = params
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, bytes)
}

func getCDPWebSocketURL(ctx context.Context, port int) (string, error) {
	urlEndpoint := fmt.Sprintf("http://127.0.0.1:%d/json", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, urlEndpoint, nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			var tabs []struct {
				WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
				Type                 string `json:"type"`
			}
			if err := json.Unmarshal(body, &tabs); err == nil {
				for _, t := range tabs {
					if t.Type == "page" && t.WebSocketDebuggerURL != "" {
						return t.WebSocketDebuggerURL, nil
					}
				}
				if len(tabs) > 0 && tabs[0].WebSocketDebuggerURL != "" {
					return tabs[0].WebSocketDebuggerURL, nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	return "", fmt.Errorf("timeout waiting for Chromium CDP port %d", port)
}

func getFreeTCPPort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 9222
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func resolveURL(baseStr, refStr string) string {
	base, err := url.Parse(baseStr)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(refStr)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}

func normalizeURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	parsed.Fragment = ""
	return parsed.String()
}
