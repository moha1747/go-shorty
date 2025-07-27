package main

import (
	"flag"
	"log"

	"github.com/moha1747/go-shorty/internal/config"
	"github.com/moha1747/go-shorty/internal/dns"
	"github.com/moha1747/go-shorty/internal/redirect"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file")
	writeDefaultConfig := flag.String("write-default-config", "", "write default config to file and exit")
	flag.Parse()

	// If requested, write default config and exit
	if *writeDefaultConfig != "" {
		if err := config.WriteDefaultConfig(*writeDefaultConfig); err != nil {
			log.Fatalf("Error writing default config: %v", err)
		}
		log.Printf("Default configuration written to %s", *writeDefaultConfig)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start services
	go dns.StartDNSServer(cfg)
	redirect.StartRedirectServer(cfg) // blocks
}
