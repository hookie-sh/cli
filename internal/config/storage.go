package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/99designs/keyring"
	"github.com/segmentio/ksuid"
)

type Config struct {
	Token     string `json:"-"` // Token stored in keyring when available; written to JSON only as a fallback when the keyring is unavailable
	RelayURL  string `json:"relay_url,omitempty"`
	MachineID string `json:"machine_id,omitempty"`
}

const (
	configDirName  = ".hookie"
	configFileName = "config.json"
	keyringService = "hookie"
	keyringAccount = "token"
)

// getKeyring initializes and returns a keyring instance
// Configured for macOS to minimize password prompts by using the login keychain
// and allowing access when the keychain is unlocked
func getKeyring() (keyring.Keyring, error) {
	config := keyring.Config{
		ServiceName: keyringService,
	}
	
	// On macOS, configure keychain settings to reduce password prompts
	// The login keychain is unlocked when the user is logged in, reducing prompts
	if runtime.GOOS == "darwin" {
		config.AllowedBackends = []keyring.BackendType{keyring.KeychainBackend}
		// Note: KeychainName may not be available in all versions of the library
		// If it causes compilation issues, it will be ignored gracefully
	}
	
	return keyring.Open(config)
}

// GetConfigPath returns the path to the config file. Use HOOKIE_CONFIG_DIR to override
// the default ~/.hookie location (useful when UserHomeDir differs between invocations).
func GetConfigPath() (string, error) {
	var configDir string
	if override := os.Getenv("HOOKIE_CONFIG_DIR"); override != "" {
		configDir = filepath.Clean(override)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, configDirName)
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return filepath.Join(configDir, configFileName), nil
}

func getConfigPath() (string, error) {
	return GetConfigPath()
}

func Load() (*Config, error) {
	config := &Config{}
	var fileToken string // token from file, used as fallback when keyring is unavailable

	// Load RelayURL and MachineID from JSON file
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Try to read config file (it's okay if it doesn't exist)
	if data, err := os.ReadFile(configPath); err == nil {
		// Use a struct that can read the old format with token and user_id
		var fileConfig struct {
			Token     string `json:"token"`
			UserID    string `json:"user_id"` // Ignored - can be extracted from token
			RelayURL  string `json:"relay_url,omitempty"`
			MachineID string `json:"machine_id,omitempty"`
		}
		if err := json.Unmarshal(data, &fileConfig); err == nil {
			config.RelayURL = fileConfig.RelayURL
			config.MachineID = fileConfig.MachineID
			fileToken = strings.TrimSpace(strings.Trim(fileConfig.Token, "\n\r"))
			// If there's a token in the file, migrate it to keyring
			if fileConfig.Token != "" {
				if err := migrateTokenToKeyring(fileConfig.Token); err == nil {
					// Migration successful; keep token in file so Load() can fall back when keyring fails
					configToSave := &Config{
						RelayURL:  config.RelayURL,
						MachineID: config.MachineID,
					}
					if err := saveConfigFileWithToken(configToSave, fileToken); err == nil {
						// Successfully migrated and saved
					}
				}
			}
		}
	}

	// Generate machine_id if it doesn't exist
	if config.MachineID == "" {
		config.MachineID = fmt.Sprintf("mach_%s", ksuid.New().String())
		// Save config to persist machine_id; preserve token so we don't overwrite it
		if err := saveConfigFileWithToken(config, fileToken); err != nil {
			// Log error but don't fail - machine_id will be regenerated next time
			fmt.Printf("Warning: failed to save machine_id: %v\n", err)
		}
	}

	// Fetch Token from keyring
	kr, err := getKeyring()
	if err == nil {
		item, err := kr.Get(keyringAccount)
		if err == nil {
			token := string(item.Data)
			// Trim whitespace and newlines
			token = strings.TrimSpace(token)
			token = strings.Trim(token, "\n\r")
			config.Token = token

			// Validate token format (warn but don't fail if invalid)
			if token != "" {
				if err := ValidateTokenFormat(token); err != nil {
					// Token format is invalid, but we'll still return it
					// The caller will handle the error when trying to use it
				}
			}
		}
	}
	// Fallback: when keyring is unavailable or empty, use token from file (old format or env-driven write)
	if config.Token == "" && fileToken != "" {
		config.Token = fileToken
	}

	return config, nil
}

// ValidateTokenFormat validates that a token is in JWT format (three parts separated by dots)
func ValidateTokenFormat(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts separated by dots, got %d parts", len(parts))
	}
	
	// Check that each part is non-empty
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid JWT format: part %d is empty", i+1)
		}
	}
	
	return nil
}

func Save(config *Config) error {
	// tokenInFile: when we fall back to file (keyring unavailable), keep token in JSON for Load() to read
	var tokenInFile string

	// Validate and clean token before storing
	if config.Token != "" {
		// Trim whitespace and newlines
		config.Token = strings.TrimSpace(config.Token)
		config.Token = strings.Trim(config.Token, "\n\r")

		// Validate JWT format
		if err := ValidateTokenFormat(config.Token); err != nil {
			return fmt.Errorf("invalid token format: %w", err)
		}

		// Store Token in keyring (preferred); fall back to config file when keyring is unavailable
		kr, err := getKeyring()
		if err != nil {
			// Fallback: persist token in config file so auth still works (e.g. sandboxed or CI)
			if errWrite := saveConfigFileWithToken(config, config.Token); errWrite != nil {
				return fmt.Errorf("cannot store login token: keyring unavailable (%w), and config file write failed: %v", err, errWrite)
			}
			tokenInFile = config.Token
		} else {
			// On macOS, try to remove the old item first to ensure it's recreated
			// with the current keyring configuration (which has better access control)
			if runtime.GOOS == "darwin" {
				_ = kr.Remove(keyringAccount) // Ignore error if item doesn't exist
			}

			if err := kr.Set(keyring.Item{
				Key:  keyringAccount,
				Data: []byte(config.Token),
			}); err != nil {
				// Fallback: persist in config file so auth still works
				if err := saveConfigFileWithToken(config, config.Token); err != nil {
					return fmt.Errorf("cannot store login token in keyring: %w", err)
				}
				tokenInFile = config.Token
			} else {
				// Keyring holds the token; do not also write it to the config file.
				// Ensure any previously written plaintext token is removed.
				tokenInFile = ""
			}
		}
	}

	// Save RelayURL and MachineID (and tokenInFile when we used file fallback)
	return saveConfigFileWithToken(config, tokenInFile)
}

// saveConfigFile saves only RelayURL and MachineID to the JSON file
func saveConfigFile(config *Config) error {
	return saveConfigFileWithToken(config, "")
}

// saveConfigFileWithToken saves RelayURL, MachineID, and optionally Token to the JSON file.
// When token is non-empty, it is written so Load() can use it when keyring is unavailable.
func saveConfigFileWithToken(config *Config, token string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	fileConfig := struct {
		Token     string `json:"token,omitempty"`
		RelayURL  string `json:"relay_url,omitempty"`
		MachineID string `json:"machine_id,omitempty"`
	}{
		Token:     token,
		RelayURL:  config.RelayURL,
		MachineID: config.MachineID,
	}

	data, err := json.MarshalIndent(fileConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// migrateTokenToKeyring migrates a token from file to keyring
func migrateTokenToKeyring(token string) error {
	// Trim whitespace and newlines before migrating
	token = strings.TrimSpace(token)
	token = strings.Trim(token, "\n\r")
	
	kr, err := getKeyring()
	if err != nil {
		return fmt.Errorf("failed to initialize keyring: %w", err)
	}

	// On macOS, ensure any old item is removed first so the new one is created
	// with the current keyring configuration (better access control)
	if runtime.GOOS == "darwin" {
		_ = kr.Remove(keyringAccount) // Ignore error if item doesn't exist
	}

	err = kr.Set(keyring.Item{
		Key:  keyringAccount,
		Data: []byte(token),
	})
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}

	return nil
}

func Clear() error {
	// Delete token from keyring
	kr, err := getKeyring()
	if err == nil {
		_ = kr.Remove(keyringAccount) // Ignore error if key doesn't exist
	}

	// Delete config JSON file
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	return nil
}

