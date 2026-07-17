//go:build !dev
// +build !dev

package relay

// RelayURL is the default relay service URL for PRODUCTION.
// This file is compiled by default (when not using -tags dev).
// With Railway, gRPC needs HTTP/2: use TCP Proxy (not HTTPS) and set HOOKIE_RELAY_URL=relay.hookie.sh:<TCP proxy port>.
// Port 443 here only works if your relay-proxy is reachable on 443 with HTTP/2 (e.g. TCP proxy forwarding to internal 443).
var RelayURL = "relay.hookie.sh"
