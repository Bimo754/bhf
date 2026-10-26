package config_test

import (
	"os"
	"testing"

	"github.com/diamond/bhf/internal/config"
)

func TestConfigLoaderDefaults(t *testing.T) {
	// Clear any overrides for standard test
	_ = os.Unsetenv("BHF_PORT")
	_ = os.Unsetenv("BHF_HOST")
	_ = os.Unsetenv("BHF_DATA_DIR")
	_ = os.Unsetenv("BHF_DB_PATH")
	_ = os.Unsetenv("BHF_LOG_LEVEL")

	cfg := config.Load()
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("expected default data dir ./data, got %s", cfg.DataDir)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("expected default log level INFO, got %s", cfg.LogLevel)
	}
}

func TestConfigLoaderEnvOverrides(t *testing.T) {
	_ = os.Setenv("BHF_PORT", "9090")
	_ = os.Setenv("BHF_HOST", "127.0.0.1")
	_ = os.Setenv("BHF_DATA_DIR", "/tmp/bhf_test_data")
	_ = os.Setenv("BHF_LOG_LEVEL", "DEBUG")

	defer func() {
		_ = os.Unsetenv("BHF_PORT")
		_ = os.Unsetenv("BHF_HOST")
		_ = os.Unsetenv("BHF_DATA_DIR")
		_ = os.Unsetenv("BHF_LOG_LEVEL")
	}()

	cfg := config.Load()
	if cfg.Port != 9090 {
		t.Errorf("expected overridden port 9090, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected overridden host 127.0.0.1, got %s", cfg.Host)
	}
	if cfg.DataDir != "/tmp/bhf_test_data" {
		t.Errorf("expected overridden data dir /tmp/bhf_test_data, got %s", cfg.DataDir)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("expected overridden log level DEBUG, got %s", cfg.LogLevel)
	}
}
