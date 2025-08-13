package redirect

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/moha1747/go-shorty/internal/config"
)

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
func StartRedirectServer(cfg *config.RedirectConfig) {
	http.HandleFunc("/", RedirectHandler(cfg.Shortcuts))

	addr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	log.Printf("HTTP server running on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
