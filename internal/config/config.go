package config

// Config holds all runtime application configuration parameters.
type Config struct {
	Port         int
	Host         string
	DatabasePath string
	LogLevel     string
	DataDir      string
}
