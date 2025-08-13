package redirect

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moha1747/go-shorty/internal/config"
)

func TestRedirectHandler(t *testing.T) {
	// Create a test shortcuts map
	testShortcuts := map[string]string{
		"test": "https://example.com",
		"go":   "https://golang.org",
	}

	// Create the handler with our test shortcuts
	handler := RedirectHandler(testShortcuts)

	tests := []struct {
		name           string
		host           string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid shortcut",
			host:           "test.u",
			expectedStatus: http.StatusFound, // 302
			expectedURL:    "https://example.com",
		},
		{
			name:           "Another valid shortcut",
			host:           "go.u",
			expectedStatus: http.StatusFound,
			expectedURL:    "https://golang.org",
		},
		{
			name:           "Invalid shortcut",
			host:           "invalid.u",
			expectedStatus: http.StatusNotFound, // 404
			expectedURL:    "",
		},
		{
			name:           "No suffix",
			host:           "test",
			expectedStatus: http.StatusFound,
			expectedURL:    "https://example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "http://"+tc.host, nil)
			req.Host = tc.host // Set the host explicitly

			// Create a recorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler
			handler(recorder, req)

			// Check the status code
			if recorder.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, recorder.Code)
			}

			// For successful redirects, check the Location header
			if tc.expectedStatus == http.StatusFound {
				location := recorder.Header().Get("Location")
				if location != tc.expectedURL {
					t.Errorf("Expected redirect to %q, got %q", tc.expectedURL, location)
				}
			}
		})
	}
}

func TestRedirectServerConfigIntegration(t *testing.T) {
	// Create a temporary directory for the test config
	tempDir, err := os.MkdirTemp("", "goshorty-redirect-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a test config file with custom shortcuts
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
dns:
  port: 5353
  upstream_dns: "8.8.8.8:53"
  local_ip: "192.168.1.1"
redirect:
  port: 8080
  address: "127.0.0.1"
  shortcuts:
    test-config: "https://test-config.example.com"
    gh-test: "https://github.com/test"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config from the file
	cfg, err := config.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the redirect config has the correct shortcuts from the file
	expectedShortcuts := map[string]string{
		"test-config": "https://test-config.example.com",
		"gh-test":     "https://github.com/test",
	}

	for shortcut, expected := range expectedShortcuts {
		actual, ok := cfg.Redirect.Shortcuts[shortcut]
		if !ok {
			t.Errorf("Expected shortcut '%s' not found in loaded config", shortcut)
		} else if actual != expected {
			t.Errorf("Expected shortcut '%s' to point to '%s', got '%s'", shortcut, expected, actual)
		}
	}

	// Test the handler with the loaded shortcuts
	handler := RedirectHandler(cfg.Redirect.Shortcuts)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup a custom HTTP client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}

	// Test each shortcut with the running server
	for shortcut, expected := range expectedShortcuts {
		t.Run("Redirect-"+shortcut, func(t *testing.T) {
			// Make a request with custom Host header
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Host = shortcut // Set the host to our shortcut

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			// Verify response
			if resp.StatusCode != http.StatusFound {
				t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
			}

			location := resp.Header.Get("Location")
			if location != expected {
				t.Errorf("Expected redirect to %q, got %q", expected, location)
			}
		})
	}

	// Test an invalid shortcut
	t.Run("Invalid-Shortcut", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Host = "nonexistent-shortcut"

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d for invalid shortcut, got %d",
				http.StatusNotFound, resp.StatusCode)
		}
	})
}
