package dns

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/moha1747/go-shorty/internal/config"
)

// DNSHandler defines the interface for handling DNS requests
type DNSHandler interface {
	HandleRequest(w dns.ResponseWriter, r *dns.Msg)
}

// DNSClient defines the interface for DNS clients
type DNSClient interface {
	Exchange(req *dns.Msg, server string) (*dns.Msg, time.Duration, error)
}

// Config holds the configuration for the DNS server
type Config struct {
	UpstreamDNS string
	LocalIP     net.IP
	Port        int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		UpstreamDNS: "1.1.1.1:53",
		LocalIP:     net.ParseIP("127.0.0.1"),
		Port:        53,
	}
}

// NewConfigFromViper creates a DNS config from the global Viper config
func NewConfigFromViper(cfg *config.Config) *Config {
	if cfg == nil {
		return DefaultConfig()
	}

	return &Config{
		UpstreamDNS: cfg.DNS.UpstreamDNS,
		LocalIP:     net.ParseIP(cfg.DNS.LocalIP),
		Port:        cfg.DNS.Port,
	}
}

// DNSServerHandler implements the DNSHandler interface
type DNSServerHandler struct {
	config *Config
	client DNSClient
}

// NewDNSServerHandler creates a new DNSServerHandler with the given configuration
func NewDNSServerHandler(config *Config, client DNSClient) *DNSServerHandler {
	if config == nil {
		config = DefaultConfig()
	}
	if client == nil {
		client = &dns.Client{}
	}
	return &DNSServerHandler{
		config: config,
		client: client,
	}
}

// StartDNSServer initializes the DNS server to handle requests for .u domains
func StartDNSServer(cfg *config.Config) {
	dnsConfig := NewConfigFromViper(cfg)
	handler := NewDNSServerHandler(dnsConfig, nil)
	dns.HandleFunc(".", handler.HandleRequest)

	port := fmt.Sprintf(":%d", dnsConfig.Port)

	go func() {
		log.Printf("DNS UDP server running on %s", port)
		if err := (&dns.Server{Addr: port, Net: "udp"}).ListenAndServe(); err != nil {
			log.Fatalf("UDP failed: %v", err)
		}
	}()
	go func() {
		log.Printf("DNS TCP server running on %s", port)
		if err := (&dns.Server{Addr: port, Net: "tcp"}).ListenAndServe(); err != nil {
			log.Fatalf("TCP failed: %v", err)
		}
	}()
}

// HandleRequest handles DNS requests, resolving .u domains locally and forwarding others
func (h *DNSServerHandler) HandleRequest(w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]
	name := q.Name

	if strings.HasSuffix(name, ".u.") {
		h.handleLocalDomain(w, r, name)
		return
	}

	h.forwardRequest(w, r, name)
}

// handleLocalDomain handles .u domain requests by returning a local IP
func (h *DNSServerHandler) handleLocalDomain(w dns.ResponseWriter, r *dns.Msg, name string) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		A: h.config.LocalIP,
	}
	msg.Answer = []dns.RR{rr}

	log.Printf("Local DNS: %s -> %s", name, h.config.LocalIP)
	_ = w.WriteMsg(msg)
}

// forwardRequest forwards non-.u domain requests to the upstream DNS server
func (h *DNSServerHandler) forwardRequest(w dns.ResponseWriter, r *dns.Msg, name string) {
	resp, _, err := h.client.Exchange(r, h.config.UpstreamDNS)
	if err != nil || resp == nil {
		log.Printf("Failed to forward %s: %v", name, err)
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeRefused
		_ = w.WriteMsg(msg)
		return
	}

	log.Printf("Forwarded DNS: %s", name)
	_ = w.WriteMsg(resp)
}
