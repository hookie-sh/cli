package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newListenFlagSet(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "listen"}
	cmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "forward" {
			return pflag.NormalizedName("forward-to")
		}
		return pflag.NormalizedName(name)
	})
	cmd.Flags().StringP("forward-to", "f", "", "Forward events to the specified endpoint URL")
	return cmd
}

func TestForwardToFlagAlias(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantVal string
		changed bool
	}{
		{
			name:    "forward alias",
			args:    []string{"--forward", "http://example.com"},
			wantVal: "http://example.com",
			changed: true,
		},
		{
			name:    "forward-to primary",
			args:    []string{"--forward-to", "http://example.com"},
			wantVal: "http://example.com",
			changed: true,
		},
		{
			name:    "forward-to shorthand",
			args:    []string{"-f", "http://example.com"},
			wantVal: "http://example.com",
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newListenFlagSet(t)
			cmd.SetArgs(tt.args)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("parse flags: %v", err)
			}

			val, err := cmd.Flags().GetString("forward-to")
			if err != nil {
				t.Fatalf("get forward-to: %v", err)
			}
			if val != tt.wantVal {
				t.Fatalf("got val %q, want %q", val, tt.wantVal)
			}
			if cmd.Flags().Changed("forward-to") != tt.changed {
				t.Fatalf("Changed(forward-to)=%v, want %v", cmd.Flags().Changed("forward-to"), tt.changed)
			}
		})
	}
}
