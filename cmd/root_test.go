package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand_help(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Hookie CLI")) {
		t.Fatalf("output missing title: %s", out.String())
	}
}

func TestResolveListenAuthentication(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		sourceID string
		appID    string
		wantAuth bool
		wantErr  bool
	}{
		{
			name:     "anonymous listen allowed when logged out",
			wantAuth: false,
		},
		{
			name:     "source listen requires login",
			sourceID: "source_123",
			wantErr:  true,
		},
		{
			name:    "app listen requires login",
			appID:   "app_123",
			wantErr: true,
		},
		{
			name:     "authenticated targeted listen allowed",
			token:    "jwt-token",
			sourceID: "source_123",
			wantAuth: true,
		},
		{
			name:     "authenticated anonymous listen stays anonymous path eligible",
			token:    "jwt-token",
			wantAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuth, err := resolveListenAuthentication(tt.token, tt.sourceID, tt.appID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveListenAuthentication() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotAuth != tt.wantAuth {
				t.Fatalf("resolveListenAuthentication() auth = %v, want %v", gotAuth, tt.wantAuth)
			}
		})
	}
}
