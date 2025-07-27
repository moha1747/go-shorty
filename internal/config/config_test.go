package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify DNS defaults
	if cfg.DNS.Port != 53 {
		t.Errorf("Expected default DNS port to be 53, got %d", cfg.DNS.Port)
	}
	if cfg.DNS.UpstreamDNS != "1.1.1.1:53" {
		t.Errorf("Expected default upstream DNS to be 1.1.1.1:53, got %s", cfg.DNS.UpstreamDNS)
	}
	if cfg.DNS.LocalIP != "127.0.0.1" {
		t.Errorf("Expected default local IP to be 127.0.0.1, got %s", cfg.DNS.LocalIP)
	}

	// Verify Redirect defaults
	if cfg.Redirect.Port != 80 {
		t.Errorf("Expected default redirect port to be 80, got %d", cfg.Redirect.Port)
	}
	if cfg.Redirect.Address != "127.0.0.1" {
		t.Errorf("Expected default redirect address to be 127.0.0.1, got %s", cfg.Redirect.Address)
	}

	// Verify shortcuts
	shortcuts := cfg.Redirect.Shortcuts
	if len(shortcuts) < 3 {
		t.Errorf("Expected at least 3 default shortcuts, got %d", len(shortcuts))
	}

	expectedShortcuts := map[string]string{
		"go": "https://go.dev",
		"gh": "https://github.com",
		"so": "https://stackoverflow.com",
	}

	for shortcut, expected := range expectedShortcuts {
		if actual, ok := shortcuts[shortcut]; !ok || actual != expected {
			t.Errorf("Expected shortcut %s to point to %s, got %s", shortcut, expected, actual)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for the test config
	tempDir, err := os.MkdirTemp("", "goshorty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
dns:
  port: 5353
  upstream_dns: "8.8.8.8:53"
  local_ip: "192.168.1.1"
redirect:
  port: 8080
  address: "0.0.0.0"
  shortcuts:
    test: "https://test.example.com"
    custom: "https://custom.example.com"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify DNS config
	if cfg.DNS.Port != 5353 {
		t.Errorf("Expected DNS port to be 5353, got %d", cfg.DNS.Port)
	}
	if cfg.DNS.UpstreamDNS != "8.8.8.8:53" {
		t.Errorf("Expected upstream DNS to be 8.8.8.8:53, got %s", cfg.DNS.UpstreamDNS)
	}
	if cfg.DNS.LocalIP != "192.168.1.1" {
		t.Errorf("Expected local IP to be 192.168.1.1, got %s", cfg.DNS.LocalIP)
	}

	// Verify Redirect config
	if cfg.Redirect.Port != 8080 {
		t.Errorf("Expected redirect port to be 8080, got %d", cfg.Redirect.Port)
	}
	if cfg.Redirect.Address != "0.0.0.0" {
		t.Errorf("Expected redirect address to be 0.0.0.0, got %s", cfg.Redirect.Address)
	}

	// Verify shortcuts
	shortcuts := cfg.Redirect.Shortcuts
	expectedShortcuts := map[string]string{
		"test":   "https://test.example.com",
		"custom": "https://custom.example.com",
	}

	for shortcut, expected := range expectedShortcuts {
		if actual, ok := shortcuts[shortcut]; !ok || actual != expected {
			t.Errorf("Expected shortcut %s to point to %s, got %s", shortcut, expected, actual)
		}
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	// Create a temporary directory for the test config
	tempDir, err := os.MkdirTemp("", "goshorty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Write the default config
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := WriteDefaultConfig(configPath); err != nil {
		t.Fatalf("Failed to write default config: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Load the config and verify it matches the defaults
	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	defaultCfg := DefaultConfig()

	// Verify key settings
	if cfg.DNS.Port != defaultCfg.DNS.Port {
		t.Errorf("Expected DNS port to be %d, got %d", defaultCfg.DNS.Port, cfg.DNS.Port)
	}

	if cfg.Redirect.Port != defaultCfg.Redirect.Port {
		t.Errorf("Expected redirect port to be %d, got %d", defaultCfg.Redirect.Port, cfg.Redirect.Port)
	}
}
