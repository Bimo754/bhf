package appintel_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/appintel"
)

func TestBuildArgs(t *testing.T) {
	cfg := appintel.BrowserConfig{
		ProxyAddr:           "127.0.0.1:8080",
		Headless:            true,
		RemoteDebuggingPort: 9222,
		WindowWidth:         1920,
		WindowHeight:        1080,
	}

	targetURL := "https://example.com/login"
	userDataDir := "/tmp/test-profile"

	args := appintel.BuildArgs(cfg, targetURL, userDataDir)
	argsStr := strings.Join(args, " ")

	expectedTokens := []string{
		"--headless=new",
		"--proxy-server=127.0.0.1:8080",
		"--user-data-dir=/tmp/test-profile",
		"--remote-debugging-port=9222",
		"--window-size=1920,1080",
		"--no-sandbox",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"https://example.com/login",
	}

	for _, token := range expectedTokens {
		if !strings.Contains(argsStr, token) {
			t.Errorf("BuildArgs missing token %q in: %s", token, argsStr)
		}
	}
}

func TestFindChromium(t *testing.T) {
	path, err := appintel.FindChromium()
	if err != nil {
		if err != appintel.ErrBrowserNotFound {
			t.Errorf("unexpected error from FindChromium: %v", err)
		}
		// Expected if host does not have chromium installed
	} else {
		if path == "" {
			t.Errorf("FindChromium returned empty path with nil error")
		}
	}
}

func TestBrowserControllerMissingBinary(t *testing.T) {
	// Point to non-existent binary to verify clean failure handling
	cfg := appintel.BrowserConfig{
		BinaryPath: "/non/existent/path/to/chrome-binary-9999",
		Timeout:    2 * time.Second,
	}

	controller := appintel.NewBrowserController(cfg)
	ctx := context.Background()

	sess, err := controller.Launch(ctx, "https://example.com")
	if err == nil {
		sess.Close()
		t.Fatal("expected error launching non-existent browser binary, got nil")
	}
}
