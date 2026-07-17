package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hookie-sh/cli/internal/config"
	"github.com/hookie-sh/cli/internal/relay"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check relay connectivity and login state",
	Long:  `Pings the Hookie relay to verify it is reachable and reports whether a login token is stored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cfgErr := config.Load()

		relayURL := os.Getenv("HOOKIE_RELAY_URL")
		if relayURL == "" {
			relayURL = relay.GetRelayURL()
		}

		client, err := relay.NewAnonymousClient(debug)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Relay: cannot connect — %v", err))
			printAuthLine(cfg, cfgErr)
			return fmt.Errorf("relay: %w", relay.WrapRelayErr(err))
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		resp, err := client.Ping(ctx)
		rtt := time.Since(start)

		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Relay: ping failed — %v", err))
			printAuthLine(cfg, cfgErr)
			errMsg := err.Error()
			if strings.Contains(errMsg, "handshake failed: EOF") || strings.Contains(errMsg, "Unavailable") {
				return fmt.Errorf("relay ping failed: %w (use relay.hookie.sh:<TCP proxy port> with TLS; or set HOOKIE_INSECURE_TLS=1 for plaintext)", err)
			}
			return fmt.Errorf("relay ping failed: %w", relay.WrapRelayErr(err))
		}

		fmt.Println(color.GreenString("Relay: OK (%s)", relayURL))
		if resp != nil && resp.ServerTimeNs != 0 {
			fmt.Printf("  %s\n", color.CyanString("RTT: %s", rtt))
		}
		printAuthLine(cfg, cfgErr)
		return nil
	},
}

func printAuthLine(cfg *config.Config, loadErr error) {
	if loadErr != nil {
		fmt.Println(color.YellowString("Auth: unknown (config error: %v)", loadErr))
		return
	}
	if strings.TrimSpace(cfg.Token) == "" {
		fmt.Println(color.YellowString("Auth: not logged in"))
		return
	}
	fmt.Println(color.GreenString("Auth: logged in"))
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
