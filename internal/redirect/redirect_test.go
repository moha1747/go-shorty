package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appconfig "github.com/moha1747/go-shorty/internal/config"
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

func TestNewRedirectConfigFromViper(t *testing.T) {
	// Test with nil config
	config := NewRedirectConfigFromViper(nil)
	if config.Address != "127.0.0.1" {
		t.Errorf("Expected default address 127.0.0.1, got %s", config.Address)
	}
	if config.Port != 80 {
		t.Errorf("Expected default port 80, got %d", config.Port)
	}
	if config.Shortcuts == nil {
		t.Error("Expected shortcuts map to be initialized")
	}
	if target, ok := config.Shortcuts["go"]; !ok || target != "https://go.dev" {
		t.Errorf("Expected shortcut 'go' to point to 'https://go.dev', got %s", target)
	}

	// Test with custom config
	customConfig := &appconfig.Config{
		Redirect: appconfig.RedirectConfig{
			Address: "0.0.0.0",
			Port:    8080,
			Shortcuts: map[string]string{
				"custom": "https://custom.example.com",
			},
		},
	}

	redirectConfig := NewRedirectConfigFromViper(customConfig)
	if redirectConfig.Address != "0.0.0.0" {
		t.Errorf("Expected custom address 0.0.0.0, got %s", redirectConfig.Address)
	}
	if redirectConfig.Port != 8080 {
		t.Errorf("Expected custom port 8080, got %d", redirectConfig.Port)
	}
	if target, ok := redirectConfig.Shortcuts["custom"]; !ok || target != "https://custom.example.com" {
		t.Errorf("Expected shortcut 'custom' to point to 'https://custom.example.com', got %s", target)
	}
}
