package redirect

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/moha1747/go-shorty/internal/config"
)

// Shortcuts maps short names to their target URLs
var Shortcuts = map[string]string{
	"go": "https://go.dev",
	"gh": "https://github.com",
	"so": "https://stackoverflow.com",
}

// RedirectConfig holds configuration for the HTTP redirect server
type RedirectConfig struct {
	Address   string
	Port      int
	Shortcuts map[string]string
}

// NewRedirectConfigFromViper creates a redirect config from the global Viper config
func NewRedirectConfigFromViper(cfg *config.Config) *RedirectConfig {
	if cfg == nil {
		return &RedirectConfig{
			Address:   "127.0.0.1",
			Port:      80,
			Shortcuts: Shortcuts,
		}
	}

	return &RedirectConfig{
		Address:   cfg.Redirect.Address,
		Port:      cfg.Redirect.Port,
		Shortcuts: cfg.Redirect.Shortcuts,
	}
}

// RedirectHandler handles HTTP requests and redirects based on hostname
func RedirectHandler(shortcuts map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := strings.TrimSuffix(r.Host, ".u")
		target, ok := shortcuts[host]
		if !ok {
			http.NotFound(w, r)
			return
		}
		log.Printf("Redirecting %s to %s", host, target)
		http.Redirect(w, r, target, http.StatusFound) // 302
	}
}

// StartRedirectServer initializes the HTTP server to handle redirects for .u domains
func StartRedirectServer(cfg *config.Config) {
	redirectCfg := NewRedirectConfigFromViper(cfg)
	http.HandleFunc("/", RedirectHandler(redirectCfg.Shortcuts))

	addr := fmt.Sprintf("%s:%d", redirectCfg.Address, redirectCfg.Port)
	log.Printf("HTTP server running on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
