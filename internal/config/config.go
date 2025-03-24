package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config holds the application configuration
type Config struct {
	Network string        // Docker network to monitor
	OutFile string        // Output file for Caddy configuration
	Notify  *NotifyConfig // Notification configuration
}

// NotifyConfig represents the notification configuration
type NotifyConfig struct {
	ContainerID string   `json:"containerId"`
	WorkingDir  string   `json:"workingDir"`
	Command     []string `json:"command"`
}

// NewConfig creates a new Config instance with values from environment variables
func NewConfig() *Config {
	return &Config{
		Network: GetEnv("CADDY_GEN_NETWORK", "gateway"),
		OutFile: GetEnv("CADDY_GEN_OUTFILE", "docker-sites.caddy"),
		Notify:  ParseNotifyConfig(GetEnv("CADDY_GEN_NOTIFY", "")),
	}
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// ParseNotifyConfig parses the notification configuration from a JSON string
func ParseNotifyConfig(raw string) *NotifyConfig {
	var config NotifyConfig

	if raw != "" {
		err := json.Unmarshal([]byte(raw), &config)
		if err != nil {
			log.Printf("Failed to parse CADDY_GEN_NOTIFY: %v", err)
		}
	}

	if len(config.Command) == 0 {
		config.Command = []string{"caddy", "reload"}
	}

	return &config
}
