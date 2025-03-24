package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")
	
	result := GetEnv("TEST_ENV_VAR", "default_value")
	if result != "test_value" {
		t.Errorf("GetEnv() = %s; want test_value", result)
	}
	
	// Test with non-existing environment variable
	result = GetEnv("NON_EXISTING_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("GetEnv() = %s; want default_value", result)
	}
}

func TestParseNotifyConfig(t *testing.T) {
	// Test with valid JSON
	validJSON := `{"containerId":"test-container","workingDir":"/app","command":["caddy","reload"]}`
	config := ParseNotifyConfig(validJSON)
	
	if config == nil {
		t.Fatal("ParseNotifyConfig() returned nil for valid JSON")
	}
	if config.ContainerID != "test-container" {
		t.Errorf("config.ContainerID = %s; want test-container", config.ContainerID)
	}
	if config.WorkingDir != "/app" {
		t.Errorf("config.WorkingDir = %s; want /app", config.WorkingDir)
	}
	if len(config.Command) != 2 || config.Command[0] != "caddy" || config.Command[1] != "reload" {
		t.Errorf("config.Command = %v; want [caddy reload]", config.Command)
	}
	
	// Test with empty string
	config = ParseNotifyConfig("")
	if config != nil {
		t.Errorf("ParseNotifyConfig() = %v; want nil", config)
	}
	
	// Test with invalid JSON
	config = ParseNotifyConfig("{invalid json}")
	if config != nil {
		t.Errorf("ParseNotifyConfig() = %v; want nil", config)
	}
}

func TestNewConfig(t *testing.T) {
	// Set environment variables
	os.Setenv("CADDY_GEN_NETWORK", "test-network")
	os.Setenv("CADDY_GEN_OUTFILE", "test-outfile")
	os.Setenv("CADDY_GEN_NOTIFY", `{"containerId":"test-container","workingDir":"/app","command":["test"]}`)
	defer func() {
		os.Unsetenv("CADDY_GEN_NETWORK")
		os.Unsetenv("CADDY_GEN_OUTFILE")
		os.Unsetenv("CADDY_GEN_NOTIFY")
	}()
	
	config := NewConfig()
	
	if config.Network != "test-network" {
		t.Errorf("config.Network = %s; want test-network", config.Network)
	}
	if config.OutFile != "test-outfile" {
		t.Errorf("config.OutFile = %s; want test-outfile", config.OutFile)
	}
	if config.Notify == nil {
		t.Fatal("config.Notify is nil")
	}
	if config.Notify.ContainerID != "test-container" {
		t.Errorf("config.Notify.ContainerID = %s; want test-container", config.Notify.ContainerID)
	}
} 
