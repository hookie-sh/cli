package relay

import (
	"os"
	"testing"
)

func TestHostFromURL(t *testing.T) {
	tests := []struct {
		in, host, port string
	}{
		{"grpc://localhost:50051", "localhost", "50051"},
		{"grpcs://example.com:443", "example.com", "443"},
		{"http://127.0.0.1:8080", "127.0.0.1", "8080"},
		{"localhost", "localhost", ""},
	}
	for _, tt := range tests {
		h, p := hostFromURL(tt.in)
		if h != tt.host || p != tt.port {
			t.Errorf("hostFromURL(%q) = %q,%q want %q,%q", tt.in, h, p, tt.host, tt.port)
		}
	}
}

func TestIsLocalhost(t *testing.T) {
	if !isLocalhost("grpc://localhost:50051") {
		t.Fatal("localhost")
	}
	if !isLocalhost("http://127.0.0.1:1") {
		t.Fatal("127.0.0.1")
	}
	if !isLocalhost("grpc://[::1]:50051") {
		t.Fatal("::1")
	}
	if isLocalhost("grpcs://relay.example.com:443") {
		t.Fatal("non-local")
	}
}

func TestNormalizeRelayURL(t *testing.T) {
	if got := normalizeRelayURL("grpc://localhost:50051"); got != "grpc://localhost:50051" {
		t.Fatalf("local unchanged: %q", got)
	}
	if got := normalizeRelayURL("grpcs://relay.hookie.sh"); got != "relay.hookie.sh:443" {
		t.Fatalf("remote no port: %q", got)
	}
	if got := normalizeRelayURL("grpcs://relay.hookie.sh:8443"); got != "grpcs://relay.hookie.sh:8443" {
		t.Fatalf("explicit port preserved: %q", got)
	}
}

func TestTransportCreds_localhostInsecure(t *testing.T) {
	t.Setenv("HOOKIE_INSECURE_TLS", "")
	c := transportCreds("grpc://127.0.0.1:50051")
	if c == nil {
		t.Fatal("nil creds")
	}
}

func TestTransportCreds_forcedInsecure(t *testing.T) {
	t.Setenv("HOOKIE_INSECURE_TLS", "1")
	t.Cleanup(func() { _ = os.Unsetenv("HOOKIE_INSECURE_TLS") })
	c := transportCreds("grpcs://relay.hookie.sh:443")
	if c == nil {
		t.Fatal("nil creds")
	}
}
