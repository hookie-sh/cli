package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hookie-sh/cli/internal/config"
	"github.com/hookie-sh/cli/internal/relay"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var sourcesCmd = &cobra.Command{
	Use:   "sources [app-id]",
	Short: "List sources for an application or all sources across all applications",
	Long:  `List all sources for a specific application, or all sources across all accessible applications if no app-id is provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var appID string
		if len(args) > 0 {
			appID = args[0]
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Run 'hookie login' first")
		}

		client, err := relay.NewClient(cfg.Token, debug)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", relay.WrapRelayErr(err))
		}
		defer client.Close()

		sources, err := client.ListSources(context.Background(), appID)
		if err != nil {
			return fmt.Errorf("failed to list sources: %w", relay.WrapRelayErr(err))
		}

		if len(sources) == 0 {
			if appID != "" {
				fmt.Printf("No sources found for application %s.\n", appID)
			} else {
				fmt.Println("No sources found.")
			}
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)

		// Include APP ID column when listing all sources (appID is empty)
		if appID == "" {
			table.Header("ID", "APP ID", "NAME", "DESCRIPTION")
		} else {
			table.Header("ID", "NAME", "DESCRIPTION")
		}

		for _, source := range sources {
			desc := source.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			if appID == "" {
				table.Append(
					color.CyanString(source.Id),
					color.YellowString(source.ApplicationId),
					source.Name,
					desc,
				)
			} else {
				table.Append(
					color.CyanString(source.Id),
					source.Name,
					desc,
				)
			}
		}
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sourcesCmd)
}
