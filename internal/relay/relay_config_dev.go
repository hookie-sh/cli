//go:build dev
// +build dev

package relay

// RelayURL is the default relay service URL for DEVELOPMENT
// This file is only compiled when building with: go build -tags dev
var RelayURL = "localhost:50051"
