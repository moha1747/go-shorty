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

// DNSServerHandler implements the DNSHandler interface
type DNSServerHandler struct {
	config *config.DNSConfig
	client DNSClient
}

// NewDNSServerHandler creates a new DNSServerHandler with the given configuration
func NewDNSServerHandler(config *config.DNSConfig, client DNSClient) *DNSServerHandler {
	if client == nil {
		client = &dns.Client{}
	}
	return &DNSServerHandler{
		config: config,
		client: client,
	}
}

// StartDNSServer initializes the DNS server to handle requests for .u domains
func StartDNSServer(cfg *config.DNSConfig) {
	handler := NewDNSServerHandler(cfg, nil)
	dns.HandleFunc(".", handler.HandleRequest)

	// Print custom extension configuration
	log.Printf("DNS using custom extension \".%s\"", cfg.Extension)

	port := fmt.Sprintf(":%d", cfg.Port)
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

	// example: ".u"
	extensionSuffix := "." + h.config.Extension

	// If request contains custom extension, point to redirect server
	if strings.HasSuffix(name, extensionSuffix) {
		h.handleLocalDomain(w, r, name)
	} else { // Otherwise send to upstream DNS
		h.forwardRequest(w, r, name)
	}
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
		A: net.ParseIP(h.config.LocalIP),
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
