package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/hookie-sh/cli/internal/config"
	"github.com/hookie-sh/cli/internal/gui"
	"github.com/hookie-sh/cli/internal/relay"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	orgID     string
	orgIDFlag string
	debug     bool
)

var rootCmd = &cobra.Command{
	Use:           "hookie",
	Short:         "Hookie CLI - Webhook event streaming tool",
	Long:          `Hookie CLI allows you to authenticate, list applications/sources, and stream webhook events in real-time.`,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Display debug information first if debug flag is set
		if debug {
			// Get command name by parsing os.Args directly
			// This is necessary because PersistentPreRun runs after flag parsing,
			// so args may not contain the subcommand name
			commandName := getCommandNameFromArgs()
			// Build command string without the full path - just use "hookie" as the command name
			commandParts := []string{"hookie"}
			if len(os.Args) > 1 {
				commandParts = append(commandParts, os.Args[1:]...)
			}
			fullCommand := strings.Join(commandParts, " ")
			printDebugInfo(commandName, orgID, fullCommand)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error: %v", err))
		os.Exit(1)
	}
}

var listenCmd = &cobra.Command{
	Use:   "listen [--forward-to <url>]",
	Short: "Listen for webhook events (anonymous or authenticated)",
	Long:  `Listen for webhook events. Without --app-id (and no app in hookie.yml), creates an anonymous ephemeral channel—even when logged in. Use --app-id to subscribe to all sources of an app, or add --source-id for a single source slug. Optionally forward events with --forward-to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forwardURL, _ := cmd.Flags().GetString("forward-to")
		forwardExplicit := cmd.Flags().Changed("forward-to")
		showUI, _ := cmd.Flags().GetBool("ui")
		showGUI := !forwardExplicit || showUI
		sourceID, _ := cmd.Flags().GetString("source-id")
		appID, _ := cmd.Flags().GetString("app-id")

		// Load repository config (if exists)
		repoConfig, _, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("failed to load repository config: %w", err)
		}

		// Store original CLI flag values to check precedence
		cliAppID := appID
		cliSourceID := sourceID

		// Priority: CLI flags > repo config
		// Use repo config values only if:
		// 1. The CLI flag for that field is empty, AND
		// 2. The conflicting CLI flag is also empty (to prevent mutual exclusion)
		if cliAppID == "" && cliSourceID == "" && repoConfig != nil && repoConfig.AppID != "" {
			appID = repoConfig.AppID
		}
		if cliSourceID == "" && repoConfig != nil && repoConfig.SourceID != "" {
			if appID == "" && repoConfig.AppID != "" {
				appID = repoConfig.AppID
			}
			sourceID = repoConfig.SourceID
		}
		if forwardURL == "" && repoConfig != nil && repoConfig.Forward != "" {
			forwardURL = repoConfig.Forward
		}

		// Build source forward map from repo config
		var sourceForwardMap map[string]*url.URL
		if repoConfig != nil && repoConfig.Sources != nil && len(repoConfig.Sources) > 0 {
			sourceForwardMap = make(map[string]*url.URL)
			for sourceID, sourceURL := range repoConfig.Sources {
				if sourceURL != "" {
					parsedURL, err := url.Parse(sourceURL)
					if err != nil {
						return fmt.Errorf("invalid forward URL for source %s: %w", sourceID, err)
					}
					if parsedURL.Scheme == "" || parsedURL.Host == "" {
						return fmt.Errorf("invalid forward URL for source %s: must include scheme and host", sourceID)
					}
					sourceForwardMap[sourceID] = parsedURL
				}
			}
		}

		// source-id (slug) requires app-id (public id)
		if sourceID != "" && appID == "" {
			return fmt.Errorf("--source-id requires --app-id (application public id)")
		}

		// Parse and validate endpoint URL if provided
		var endpointURL *url.URL
		if forwardURL != "" {
			parsedURL, err := url.Parse(forwardURL)
			if err != nil {
				return fmt.Errorf("invalid endpoint URL: %w", err)
			}
			if parsedURL.Scheme == "" || parsedURL.Host == "" {
				return fmt.Errorf("invalid endpoint URL: must include scheme and host (e.g., http://localhost:3001/webhooks)")
			}
			endpointURL = parsedURL
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		authenticated, err := resolveListenAuthentication(cfg.Token, sourceID, appID)
		if err != nil {
			return err
		}

		// Start GUI when: no --forward-to, or --forward-to + --ui
		var guiURL *url.URL
		if showGUI {
			port := gui.DefaultPort()
			var started bool
			var err error
			guiURL, started, err = gui.AcquireOrUseServer(port)
			if err != nil {
				return fmt.Errorf("GUI: %w", err)
			}
			if started {
				fmt.Println(color.CyanString("GUI available at %s/", guiURL.String()))
				openBrowserTo(guiURL.String())
			} else {
				fmt.Println(color.CyanString("Using existing GUI at %s/", guiURL.String()))
			}
		}

		if !authenticated {
			return runAnonymousListen(endpointURL, guiURL, false)
		}

		if sourceID == "" && appID == "" {
			return runAnonymousListen(endpointURL, guiURL, true)
		}

		client, err := relay.NewClient(cfg.Token, debug)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", relay.WrapRelayErr(err))
		}
		defer client.Close()

		effectiveOrgID := orgID
		if sourceID != "" {
			return runListen(sourceID, appID, effectiveOrgID, endpointURL, sourceForwardMap, guiURL)
		}
		return runListen("", appID, effectiveOrgID, endpointURL, sourceForwardMap, guiURL)
	},
}

func resolveListenAuthentication(token, sourceID, appID string) (bool, error) {
	token = strings.TrimSpace(token)
	targetedListen := sourceID != "" || appID != ""
	if token == "" && targetedListen {
		return false, fmt.Errorf("not authenticated. Run 'hookie login' first")
	}

	return token != "", nil
}

// getCommandNameFromArgs extracts the subcommand name from os.Args
// It skips the program name and any flags to find the actual command
func getCommandNameFromArgs() string {
	if len(os.Args) < 2 {
		return "root"
	}

	// Skip program name (os.Args[0])
	// Look for the first non-flag argument
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Skip flags (starting with -)
		if strings.HasPrefix(arg, "-") {
			// Skip flag value if it's not a flag itself
			if strings.Contains(arg, "=") {
				continue
			}
			// Check if next arg is a value (not a flag)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				i++ // Skip the flag value
			}
			continue
		}
		// Found a non-flag argument - this should be the command name
		return arg
	}

	return "root"
}

func openBrowserTo(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}

func init() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%w\n\n%s", err, c.UsageString())
	})
	rootCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "Organization ID (can be set globally or per command)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show detailed information (headers, query params, body, etc.)")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version")
	listenCmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "forward" {
			return pflag.NormalizedName("forward-to")
		}
		return pflag.NormalizedName(name)
	})
	listenCmd.Flags().StringP("forward-to", "f", "", "Forward events to the specified endpoint URL")
	listenCmd.Flags().Bool("ui", false, "Show local UI when forwarding with --forward-to")
	listenCmd.Flags().StringP("source-id", "s", "", "Subscribe to a specific source")
	listenCmd.Flags().StringP("app-id", "a", "", "Subscribe to all sources of an application")
	rootCmd.AddCommand(listenCmd)
}
