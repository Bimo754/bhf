package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AppIntelHandler manages browser exploration sessions, interceptor proxy, and autonomous crawling.
type AppIntelHandler struct {
	trafficRepo     *database.TrafficRepo
	targetRepo      *database.TargetRepo
	scanRepo        *database.ScanRepo
	mu              sync.Mutex
	activeProxy     *appintel.ProxyServer
	proxyTarget     string
	proxyAddr       string
	activeExplorers map[string]*appintel.AutoExplorer
}

// NewAppIntelHandler creates a new AppIntelHandler.
func NewAppIntelHandler(
	trafficRepo *database.TrafficRepo,
	targetRepo *database.TargetRepo,
	scanRepo *database.ScanRepo,
) *AppIntelHandler {
	return &AppIntelHandler{
		trafficRepo:     trafficRepo,
		targetRepo:      targetRepo,
		scanRepo:        scanRepo,
		activeExplorers: make(map[string]*appintel.AutoExplorer),
	}
}

// ProxyStatusResponse describes the proxy server state.
type ProxyStatusResponse struct {
	Running  bool   `json:"running"`
	Addr     string `json:"addr,omitempty"`
	TargetID string `json:"target_id,omitempty"`
}

// GetProxyStatus returns whether the interceptor proxy is currently active.
func (h *AppIntelHandler) GetProxyStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	running := h.activeProxy != nil && h.activeProxy.Addr() != ""
	JSON(w, http.StatusOK, ProxyStatusResponse{
		Running:  running,
		Addr:     h.proxyAddr,
		TargetID: h.proxyTarget,
	})
}

// StartProxy starts the HTTP/HTTPS interceptor proxy on the requested or default port.
func (h *AppIntelHandler) StartProxy(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TargetID string `json:"target_id"`
		Port     int    `json:"port"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.TargetID == "" {
		ErrorJSON(w, http.StatusBadRequest, "target_id is required")
		return
	}

	target, err := h.targetRepo.GetByID(body.TargetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.activeProxy != nil && h.activeProxy.Addr() != "" {
		JSON(w, http.StatusOK, ProxyStatusResponse{
			Running:  true,
			Addr:     h.proxyAddr,
			TargetID: h.proxyTarget,
		})
		return
	}

	port := 8888
	if body.Port > 0 && body.Port <= 65535 {
		port = body.Port
	}

	cfg := appintel.InterceptorConfig{
		TargetID:     target.ID,
		TargetDomain: target.Domain,
		ScanID:       fmt.Sprintf("proxy-%d", time.Now().Unix()),
		TrafficRepo:  h.trafficRepo,
	}

	proxy := appintel.NewProxyServer(cfg)
	addr, err := proxy.Start(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		// Fallback to dynamic port if requested port is busy
		addr, err = proxy.Start("127.0.0.1:0")
		if err != nil {
			ErrorJSON(w, http.StatusInternalServerError, "failed to start proxy server: "+err.Error())
			return
		}
	}

	h.activeProxy = proxy
	h.proxyTarget = target.ID
	h.proxyAddr = addr

	JSON(w, http.StatusOK, ProxyStatusResponse{
		Running:  true,
		Addr:     addr,
		TargetID: target.ID,
	})
}

// StopProxy halts the active interceptor proxy server.
func (h *AppIntelHandler) StopProxy(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.activeProxy != nil {
		_ = h.activeProxy.Stop()
		h.activeProxy = nil
		h.proxyTarget = ""
		h.proxyAddr = ""
	}

	JSON(w, http.StatusOK, ProxyStatusResponse{
		Running: false,
	})
}

// LaunchBrowser launches an isolated browser session directed at the target scope.
func (h *AppIntelHandler) LaunchBrowser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TargetID string `json:"target_id"`
		URL      string `json:"url"`
		Headless bool   `json:"headless"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	target, err := h.targetRepo.GetByID(body.TargetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	targetURL := body.URL
	if targetURL == "" {
		targetURL = "https://" + target.Domain
	}

	h.mu.Lock()
	proxyAddr := h.proxyAddr
	h.mu.Unlock()

	browserCfg := appintel.BrowserConfig{
		Headless:  body.Headless,
		ProxyAddr: proxyAddr,
		Timeout:   30 * time.Second,
	}

	controller := appintel.NewBrowserController(browserCfg)
	sessionID := "browser-" + strconv.FormatInt(time.Now().Unix(), 10)

	session, err := controller.Launch(r.Context(), targetURL)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to launch browser: "+err.Error())
		return
	}

	// Keep session alive briefly or close when context ends
	go func() {
		time.Sleep(10 * time.Second)
		_ = session.Close()
	}()

	JSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID,
		"target_url": targetURL,
		"proxy_addr": proxyAddr,
		"message":    "Browser session initiated successfully.",
	})
}

// AutoExploreRequest configures autonomous interaction and site discovery.
type AutoExploreRequest struct {
	StartURL     string `json:"start_url"`
	MaxPages     int    `json:"max_pages"`
	MaxDepth     int    `json:"max_depth"`
	PreferChrome bool   `json:"prefer_chrome"`
}

// AutoExplore executes autonomous site discovery, DOM interaction, and traffic interception.
func (h *AppIntelHandler) AutoExplore(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	target, err := h.targetRepo.GetByID(targetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	var req AutoExploreRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.MaxPages <= 0 {
		req.MaxPages = 25
	}
	if req.MaxDepth <= 0 {
		req.MaxDepth = 3
	}

	now := time.Now()
	scan := &models.Scan{
		ID:        uuid.New().String(),
		TargetID:  targetID,
		ToolName:  "auto_explorer",
		Status:    models.StatusRunning,
		StartedAt: now,
		Arguments: []string{req.StartURL},
	}
	if h.scanRepo != nil {
		_ = h.scanRepo.Create(scan)
	}

	cfg := appintel.AutoExplorerConfig{
		Target:       target,
		TrafficRepo:  h.trafficRepo,
		ScanID:       scan.ID,
		StartURL:     req.StartURL,
		MaxPages:     req.MaxPages,
		MaxDepth:     req.MaxDepth,
		PreferChrome: req.PreferChrome,
	}

	explorer := appintel.NewAutoExplorer(cfg, nil)

	h.mu.Lock()
	h.activeExplorers[targetID] = explorer
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.activeExplorers, targetID)
		h.mu.Unlock()
	}()

	progress, err := explorer.Run(r.Context())
	completedAt := time.Now()
	if h.scanRepo != nil {
		if err != nil {
			_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusFailed, &completedAt, err.Error())
		} else {
			_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusCompleted, &completedAt, "")
		}
	}

	if err != nil && progress == nil {
		ErrorJSON(w, http.StatusInternalServerError, "auto explorer failed: "+err.Error())
		return
	}

	JSON(w, http.StatusOK, progress)
}

// GetAutoExploreStatus returns progress for an ongoing or last-run exploration.
func (h *AppIntelHandler) GetAutoExploreStatus(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")

	h.mu.Lock()
	explorer, running := h.activeExplorers[targetID]
	h.mu.Unlock()

	if !running || explorer == nil {
		JSON(w, http.StatusOK, map[string]any{
			"target_id": targetID,
			"running":   false,
		})
		return
	}

	progress := explorer.GetProgress()
	JSON(w, http.StatusOK, map[string]any{
		"target_id": targetID,
		"running":   true,
		"progress":  progress,
	})
}
