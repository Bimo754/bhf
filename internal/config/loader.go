package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	// Ensure user go bin and local bin are always available in PATH for tool discovery
	curPath := os.Getenv("PATH")
	homeDir, _ := os.UserHomeDir()
	goBin := filepath.Join(homeDir, "go", "bin")
	if !strings.Contains(curPath, goBin) {
		_ = os.Setenv("PATH", curPath+":"+goBin+":/usr/local/bin")
	}
	port := 8080
	if val := os.Getenv("BHF_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil && p > 0 {
			port = p
		}
	}

	host := "0.0.0.0"
	if val := os.Getenv("BHF_HOST"); val != "" {
		host = val
	}

	dataDir := "./data"
	if val := os.Getenv("BHF_DATA_DIR"); val != "" {
		dataDir = val
	}

	// Ensure data directory exists
	_ = os.MkdirAll(dataDir, 0755)

	dbPath := filepath.Join(dataDir, "bhf.db")
	if val := os.Getenv("BHF_DB_PATH"); val != "" {
		dbPath = val
	}

	logLevel := "INFO"
	if val := os.Getenv("BHF_LOG_LEVEL"); val != "" {
		logLevel = val
	}

	return &Config{
		Port:         port,
		Host:         host,
		DataDir:      dataDir,
		DatabasePath: dbPath,
		LogLevel:     logLevel,
	}
}
