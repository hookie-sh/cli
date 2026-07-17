package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTokenFormat(t *testing.T) {
	if err := ValidateTokenFormat(""); err == nil {
		t.Fatal("empty token should error")
	}
	if err := ValidateTokenFormat("a.b"); err == nil {
		t.Fatal("two parts should error")
	}
	if err := ValidateTokenFormat("a.b.c"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTokenFormat("a..c"); err == nil {
		t.Fatal("empty middle part should error")
	}
}

func TestGetConfigPath_usesHookieConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOOKIE_CONFIG_DIR", dir)
	path, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	wantName := filepath.Join(dir, configFileName)
	if filepath.Clean(path) != filepath.Clean(wantName) {
		t.Fatalf("path = %q want base %q", path, wantName)
	}
	if !strings.HasSuffix(path, configFileName) {
		t.Fatalf("path = %q", path)
	}
}

// TestSaveConfigFileWithToken_EmptyTokenOmittedFromJSON verifies that when the
// keyring holds the token (the success path in Save), passing an empty token
// to saveConfigFileWithToken results in a config file with no "token" field
// (via the `omitempty` tag), and that the write truncates the file rather than
// merging with a previously written token.
//
// Note: getKeyring() opens a real OS keyring backend and is not injectable, so
// the keyring-success / keyring-unavailable branches of Save() itself cannot be
// exercised in this unit test; those are covered by manual/integration testing.
// This test only covers the file-writing behavior that both branches rely on.
func TestSaveConfigFileWithToken_EmptyTokenOmittedFromJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOOKIE_CONFIG_DIR", dir)

	cfg := &Config{RelayURL: "https://relay.example.com", MachineID: "mach_test"}

	// Simulate a stale plaintext token already on disk from a previous version.
	if err := saveConfigFileWithToken(cfg, "stale.token.value"); err != nil {
		t.Fatalf("seed write failed: %v", err)
	}

	// Now simulate the keyring-success path: Save() calls this with tokenInFile == "".
	if err := saveConfigFileWithToken(cfg, ""); err != nil {
		t.Fatalf("empty-token write failed: %v", err)
	}

	path, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["token"]; ok {
		t.Fatalf("expected no token field in config file, got: %s", data)
	}
	if raw["relay_url"] != "https://relay.example.com" {
		t.Fatalf("relay_url not preserved: %s", data)
	}

	// Sanity check the other direction: a non-empty token IS written (this is
	// the genuine fallback path used when the keyring is unavailable).
	if err := saveConfigFileWithToken(cfg, "fresh.token.value"); err != nil {
		t.Fatalf("non-empty-token write failed: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["token"] != "fresh.token.value" {
		t.Fatalf("expected token field to be written in fallback path, got: %s", data)
	}
}
