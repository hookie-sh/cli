package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hookie-sh/cli/internal/relay"
	"github.com/spf13/cobra"
)

var relayTestCmd = &cobra.Command{
	Use:   "relay-test",
	Short: "Test connectivity to the relay (hidden)",
	RunE: func(cmd *cobra.Command, args []string) error {
		relayURL := os.Getenv("HOOKIE_RELAY_URL")
		if relayURL == "" {
			relayURL = relay.GetRelayURL()
		}
		if relayURL == "" {
			return fmt.Errorf("relay URL not configured: set HOOKIE_RELAY_URL or use default")
		}

		client, err := relay.NewAnonymousClient(debug)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		resp, err := client.Ping(ctx)
		rtt := time.Since(start)

		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "handshake failed: EOF") || strings.Contains(errMsg, "Unavailable") {
				return fmt.Errorf("ping failed: %w (use relay.hookie.sh:<TCP proxy port> with TLS; or set HOOKIE_INSECURE_TLS=1 for plaintext)", err)
			}
			return fmt.Errorf("ping failed: %w", relay.WrapRelayErr(err))
		}

		fmt.Println(color.GreenString("OK %s", relayURL))
		if resp != nil && resp.ServerTimeNs != 0 {
			fmt.Printf("  RTT: %s\n", rtt)
		}
		return nil
	},
}

func init() {
	relayTestCmd.Hidden = true
	rootCmd.AddCommand(relayTestCmd)
}
