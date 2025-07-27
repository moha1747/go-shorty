package dns

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// MockDNSResponseWriter implements the dns.ResponseWriter interface for testing
type MockDNSResponseWriter struct {
	lastMsg *dns.Msg
}

// Network returns the network type for the mock response writer.
func (m *MockDNSResponseWriter) Network() string {
	return "udp"
}

func (m *MockDNSResponseWriter) LocalAddr() net.Addr {
	return nil
}

func (m *MockDNSResponseWriter) RemoteAddr() net.Addr {
	return nil
}

func (m *MockDNSResponseWriter) WriteMsg(msg *dns.Msg) error {
	m.lastMsg = msg
	return nil
}

func (m *MockDNSResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *MockDNSResponseWriter) Close() error {
	return nil
}

func (m *MockDNSResponseWriter) TsigStatus() error {
	return nil
}

func (m *MockDNSResponseWriter) TsigTimersOnly(bool) {}

func (m *MockDNSResponseWriter) Hijack() {}

// MockDNSClient implements a mock dns client for testing
type MockDNSClient struct {
	shouldFail  bool
	mockResp    *dns.Msg
	lastRequest *dns.Msg
	lastServer  string
}

// Exchange mocks the Exchange method of dns.Client
func (m *MockDNSClient) Exchange(req *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
	m.lastRequest = req
	m.lastServer = server

	if m.shouldFail {
		return nil, 0, fmt.Errorf("mock DNS exchange error")
	}

	return m.mockResp, 0, nil
}

func TestHandleLocalDomain(t *testing.T) {
	// Create a new handler with a test configuration
	config := &Config{
		UpstreamDNS: "8.8.8.8:53",
		LocalIP:     net.ParseIP("192.168.1.1"),
		Port:        53,
	}
	handler := NewDNSServerHandler(config, nil)

	// Create a test request for a .u domain
	req := new(dns.Msg)
	req.SetQuestion("test.u.", dns.TypeA)

	// Create a mock writer to capture the response
	mockWriter := &MockDNSResponseWriter{}

	// Call the handler
	handler.HandleRequest(mockWriter, req)

	// Verify the response
	if mockWriter.lastMsg == nil {
		t.Fatalf("Expected a response message, got nil")
	}

	if len(mockWriter.lastMsg.Answer) != 1 {
		t.Fatalf("Expected 1 answer, got %d", len(mockWriter.lastMsg.Answer))
	}

	// Check that the response contains our configured local IP
	answer, ok := mockWriter.lastMsg.Answer[0].(*dns.A)
	if !ok {
		t.Fatalf("Expected answer of type *dns.A, got %T", mockWriter.lastMsg.Answer[0])
	}

	if !answer.A.Equal(config.LocalIP) {
		t.Errorf("Expected IP %v, got %v", config.LocalIP, answer.A)
	}
}

func TestForwardRequest(t *testing.T) {
	// Create a mock response
	mockResp := new(dns.Msg)
	mockResp.SetReply(&dns.Msg{})

	// Create a mock client
	mockClient := &MockDNSClient{
		shouldFail: false,
		mockResp:   mockResp,
	}

	// Create a handler with the mock client
	config := &Config{UpstreamDNS: "8.8.8.8:53", Port: 53}
	handler := NewDNSServerHandler(config, mockClient)

	// Create a test request for a non-.u domain
	req := new(dns.Msg)
	req.SetQuestion("example.com.", dns.TypeA)

	// Create a mock writer to capture the response
	mockWriter := &MockDNSResponseWriter{}

	// Call the handler
	handler.HandleRequest(mockWriter, req)

	// Verify the client was called with the correct server
	if mockClient.lastServer != config.UpstreamDNS {
		t.Errorf("Expected client to use server %s, got %s", config.UpstreamDNS, mockClient.lastServer)
	}

	// Verify the response was forwarded correctly
	if mockWriter.lastMsg != mockResp {
		t.Errorf("Expected forwarded response to match mock response")
	}
}

func TestForwardRequestFailure(t *testing.T) {
	// Create a mock client that fails
	mockClient := &MockDNSClient{
		shouldFail: true,
	}

	// Create a handler with the mock client
	config := &Config{UpstreamDNS: "8.8.8.8:53", Port: 53}
	handler := NewDNSServerHandler(config, mockClient)

	// Create a test request
	req := new(dns.Msg)
	req.SetQuestion("example.com.", dns.TypeA)

	// Create a mock writer to capture the response
	mockWriter := &MockDNSResponseWriter{}

	// Call the handler
	handler.HandleRequest(mockWriter, req)

	// Verify that a failure response was generated
	if mockWriter.lastMsg == nil {
		t.Fatalf("Expected a response message on failure, got nil")
	}

	if mockWriter.lastMsg.Rcode != dns.RcodeRefused {
		t.Errorf("Expected Rcode %d, got %d", dns.RcodeRefused, mockWriter.lastMsg.Rcode)
	}
}
