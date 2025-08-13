package main

import (
	"log"

	"github.com/moha1747/go-shorty/internal/config"
	"github.com/moha1747/go-shorty/internal/dns"
	"github.com/moha1747/go-shorty/internal/redirect"
)

const (
	configPath = "./config/config.yaml"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start services
	go dns.StartDNSServer(&cfg.DNS)
	redirect.StartRedirectServer(&cfg.Redirect) // blocks
}
