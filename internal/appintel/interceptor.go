package appintel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
	"github.com/google/uuid"
)

// InterceptorConfig controls behavior of the traffic interceptor and proxy.
type InterceptorConfig struct {
	TargetID     string
	TargetDomain string
	ScanID       string
	TrafficRepo  *database.TrafficRepo
	MaxBodyBytes int64 // default 1MB
}

// RecordingTransport is an http.RoundTripper that captures request/response transactions.
type RecordingTransport struct {
	Base   http.RoundTripper
	Config InterceptorConfig
}

// NewRecordingTransport creates a new RecordingTransport wrapping base transport.
func NewRecordingTransport(cfg InterceptorConfig, base http.RoundTripper) *RecordingTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 1024 * 1024 // 1 MB
	}
	return &RecordingTransport{
		Base:   base,
		Config: cfg,
	}
}

// RoundTrip executes the HTTP request and captures the full transaction into SQLite if in-scope.
func (t *RecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqURL := req.URL.String()
	host := req.URL.Hostname()
	if host == "" {
		host = req.Host
	}

	// Scope verification: only record in-scope traffic if TargetDomain is specified
	inScope := true
	if t.Config.TargetDomain != "" {
		inScope = tools.IsInScope(host, t.Config.TargetDomain)
	}

	// Capture request body safely
	var reqBodyBytes []byte
	if req.Body != nil && inScope {
		var buf bytes.Buffer
		tee := io.TeeReader(req.Body, &buf)
		reqBodyBytes, _ = io.ReadAll(io.LimitReader(tee, t.Config.MaxBodyBytes))
		req.Body = io.NopCloser(io.MultiReader(&buf, req.Body))
	}

	start := time.Now()
	resp, err := t.Base.RoundTrip(req)
	duration := time.Since(start).Milliseconds()

	if err != nil || !inScope || t.Config.TrafficRepo == nil {
		return resp, err
	}

	// Capture response body safely
	var resBodyBytes []byte
	if resp != nil && resp.Body != nil {
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)
		resBodyBytes, _ = io.ReadAll(io.LimitReader(tee, t.Config.MaxBodyBytes))
		resp.Body = io.NopCloser(io.MultiReader(&buf, resp.Body))
	}

	// Serialize headers
	reqHeadersJSON, _ := json.Marshal(req.Header)
	var resHeadersJSON []byte
	var statusCode int
	var contentType string
	var contentLength int64

	if resp != nil {
		statusCode = resp.StatusCode
		contentType = resp.Header.Get("Content-Type")
		contentLength = resp.ContentLength
		if contentLength < 0 {
			contentLength = int64(len(resBodyBytes))
		}
		resHeadersJSON, _ = json.Marshal(resp.Header)
	}

	record := &models.TrafficRecord{
		ID:                 uuid.New().String(),
		TargetID:           t.Config.TargetID,
		ScanID:             t.Config.ScanID,
		Method:             req.Method,
		URL:                reqURL,
		Path:               req.URL.Path,
		Host:               host,
		RequestHeaders:     string(reqHeadersJSON),
		RequestBody:        string(reqBodyBytes),
		ResponseStatusCode: statusCode,
		ResponseHeaders:    string(resHeadersJSON),
		ResponseBody:       string(resBodyBytes),
		ContentType:        contentType,
		ContentLength:      contentLength,
		DurationMs:         duration,
		Category:           inferCategory(req.Method, req.URL.Path, contentType),
		AuthContext:        inferAuthContext(req.Header),
		Parameters:         extractParameters(req.URL, string(reqBodyBytes), contentType),
		CreatedAt:          time.Now().UTC(),
	}

	_ = t.Config.TrafficRepo.Insert(record)

	return resp, nil
}

// ProxyServer runs an HTTP/HTTPS proxy listener that records traffic through RecordingTransport.
type ProxyServer struct {
	server   *http.Server
	listener net.Listener
	config   InterceptorConfig
	mu       sync.Mutex
	running  bool
}

// NewProxyServer creates a proxy server on 127.0.0.1 with dynamic or specified port.
func NewProxyServer(cfg InterceptorConfig) *ProxyServer {
	return &ProxyServer{
		config: cfg,
	}
}

// Start begins listening on the specified or dynamic port (e.g., "127.0.0.1:0").
func (p *ProxyServer) Start(addr string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return p.listener.Addr().String(), nil
	}

	if addr == "" {
		addr = "127.0.0.1:0"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to bind proxy listener: %w", err)
	}
	p.listener = listener

	transport := NewRecordingTransport(p.config, http.DefaultTransport)

	p.server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				p.handleHTTPSConnect(w, r)
				return
			}
			p.handleHTTP(w, r, transport)
		}),
	}

	p.running = true
	go func() {
		_ = p.server.Serve(p.listener)
	}()

	return p.listener.Addr().String(), nil
}

// Addr returns the listening address of the proxy server.
func (p *ProxyServer) Addr() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listener != nil {
		return p.listener.Addr().String()
	}
	return ""
}

// Stop shuts down the proxy server.
func (p *ProxyServer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}
	p.running = false
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

func (p *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request, transport *RecordingTransport) {
	outReq := r.Clone(r.Context())
	if outReq.URL.Scheme == "" {
		outReq.URL.Scheme = "http"
	}
	if outReq.URL.Host == "" {
		outReq.URL.Host = r.Host
	}
	outReq.RequestURI = ""

	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (p *ProxyServer) handleHTTPSConnect(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	_, _ = io.Copy(destination, source)
}

// inferCategory assigns an initial functional category based on method, path, and content type.
func inferCategory(method, path, contentType string) string {
	lowerPath := strings.ToLower(path)
	if strings.Contains(lowerPath, "login") || strings.Contains(lowerPath, "auth") ||
		strings.Contains(lowerPath, "logout") || strings.Contains(lowerPath, "signin") ||
		strings.Contains(lowerPath, "signup") || strings.Contains(lowerPath, "register") ||
		strings.Contains(lowerPath, "password") || strings.Contains(lowerPath, "oauth") {
		return "auth"
	}
	if strings.Contains(lowerPath, "upload") || strings.Contains(lowerPath, "avatar") ||
		strings.Contains(lowerPath, "attachment") || strings.Contains(lowerPath, "import") ||
		strings.Contains(contentType, "multipart/form-data") {
		return "upload"
	}
	if strings.Contains(lowerPath, "admin") || strings.Contains(lowerPath, "dashboard") ||
		strings.Contains(lowerPath, "metrics") || strings.Contains(lowerPath, "actuator") ||
		strings.Contains(lowerPath, "manage") {
		return "admin"
	}
	if strings.Contains(lowerPath, "search") || strings.Contains(lowerPath, "query") ||
		strings.Contains(lowerPath, "filter") || strings.Contains(lowerPath, "find") {
		return "search"
	}
	if strings.Contains(lowerPath, "/api/") || strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete {
		return "api"
	}
	return "general"
}

// inferAuthContext detects presence of auth tokens, cookies, or api keys.
func inferAuthContext(headers http.Header) string {
	if auth := headers.Get("Authorization"); auth != "" {
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return "bearer_token"
		}
		return "auth_header"
	}
	if cookie := headers.Get("Cookie"); cookie != "" {
		if strings.Contains(strings.ToLower(cookie), "sess") || strings.Contains(strings.ToLower(cookie), "token") {
			return "session_cookie"
		}
		return "cookie"
	}
	if headers.Get("X-API-Key") != "" || headers.Get("X-Api-Key") != "" {
		return "api_key"
	}
	return "unauthenticated"
}

// extractParameters extracts query and body parameter names into a JSON string.
func extractParameters(u *url.URL, body string, contentType string) string {
	type paramInfo struct {
		Name string `json:"name"`
		Type string `json:"type"` // query, body, json
	}
	params := make([]paramInfo, 0)

	if u != nil {
		for key := range u.Query() {
			params = append(params, paramInfo{Name: key, Type: "query"})
		}
	}

	if strings.Contains(contentType, "application/json") && body != "" {
		var jsonMap map[string]any
		if err := json.Unmarshal([]byte(body), &jsonMap); err == nil {
			for key := range jsonMap {
				params = append(params, paramInfo{Name: key, Type: "json"})
			}
		}
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") && body != "" {
		if values, err := url.ParseQuery(body); err == nil {
			for key := range values {
				params = append(params, paramInfo{Name: key, Type: "body"})
			}
		}
	}

	bytes, _ := json.Marshal(params)
	return string(bytes)
}
