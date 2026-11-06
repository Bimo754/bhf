package appintel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Common errors for browser exploration.
var (
	ErrBrowserNotFound = errors.New("chromium or google-chrome executable not found in system PATH")
)

// BrowserConfig specifies configuration options for the headless browser session.
type BrowserConfig struct {
	BinaryPath          string
	ProxyAddr           string
	Headless            bool
	UserDataDir         string
	RemoteDebuggingPort int
	WindowWidth         int
	WindowHeight        int
	Timeout             time.Duration
}

// BrowserSession represents an active isolated browser execution instance.
type BrowserSession struct {
	cmd         *exec.Cmd
	userDataDir string
	targetURL   string
	proxyAddr   string
	cancel      context.CancelFunc
}

// BrowserController manages the lifecycle of isolated browser exploration instances.
type BrowserController struct {
	config BrowserConfig
}

// NewBrowserController creates a new BrowserController.
func NewBrowserController(cfg BrowserConfig) *BrowserController {
	if cfg.WindowWidth <= 0 {
		cfg.WindowWidth = 1920
	}
	if cfg.WindowHeight <= 0 {
		cfg.WindowHeight = 1080
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}
	return &BrowserController{config: cfg}
}

// FindChromium searches standard binary names and system paths for a Chromium-compatible browser.
func FindChromium() (string, error) {
	candidates := []string{
		"chromium",
		"google-chrome",
		"google-chrome-stable",
		"chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium-browser",
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", ErrBrowserNotFound
}

// BuildArgs constructs the command-line arguments for Chrome/Chromium.
func BuildArgs(cfg BrowserConfig, targetURL string, userDataDir string) []string {
	args := []string{
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-software-rasterizer",
		"--disable-background-networking",
		"--disable-default-apps",
		"--disable-extensions",
		"--disable-sync",
		"--no-first-run",
		"--mute-audio",
		fmt.Sprintf("--window-size=%d,%d", cfg.WindowWidth, cfg.WindowHeight),
	}

	if cfg.Headless {
		args = append(args, "--headless=new")
	}

	if userDataDir != "" {
		args = append(args, fmt.Sprintf("--user-data-dir=%s", userDataDir))
	}

	if cfg.ProxyAddr != "" {
		args = append(args, fmt.Sprintf("--proxy-server=%s", cfg.ProxyAddr))
		// Trust proxy / ignore cert errors for local MITM inspection in controlled research
		args = append(args, "--ignore-certificate-errors")
	}

	if cfg.RemoteDebuggingPort > 0 {
		args = append(args, fmt.Sprintf("--remote-debugging-port=%d", cfg.RemoteDebuggingPort))
	}

	if targetURL != "" {
		args = append(args, targetURL)
	}

	return args
}

// Launch starts an isolated browser session targeting a URL.
func (c *BrowserController) Launch(ctx context.Context, targetURL string) (*BrowserSession, error) {
	binary := c.config.BinaryPath
	if binary == "" {
		found, err := FindChromium()
		if err != nil {
			return nil, err
		}
		binary = found
	}

	userDataDir := c.config.UserDataDir
	isTempDir := false
	if userDataDir == "" {
		tempDir, err := os.MkdirTemp("", "bhf-browser-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create browser user data dir: %w", err)
		}
		userDataDir = tempDir
		isTempDir = true
	}

	args := BuildArgs(c.config, targetURL, userDataDir)

	sessCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	cmd := exec.CommandContext(sessCtx, binary, args...)
	// Process group isolation for clean subprocess tree termination
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		cancel()
		if isTempDir {
			_ = os.RemoveAll(userDataDir)
		}
		return nil, fmt.Errorf("failed to start browser process: %w", err)
	}

	return &BrowserSession{
		cmd:         cmd,
		userDataDir: userDataDir,
		targetURL:   targetURL,
		proxyAddr:   c.config.ProxyAddr,
		cancel:      cancel,
	}, nil
}

// Close terminates the browser process group and cleans up temporary user data.
func (s *BrowserSession) Close() error {
	s.cancel()

	if s.cmd != nil && s.cmd.Process != nil {
		// Terminate process group cleanly
		_ = syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
		_ = s.cmd.Wait()
	}

	if s.userDataDir != "" {
		_ = os.RemoveAll(s.userDataDir)
	}

	return nil
}

// PID returns the process ID of the running browser.
func (s *BrowserSession) PID() int {
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Pid
	}
	return 0
}
