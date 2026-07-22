package cmd

import "testing"

func TestValidateListenSelectors(t *testing.T) {
	tests := []struct {
		name     string
		appID    string
		sourceID string
		wantErr  bool
	}{
		{name: "app only", appID: "billing-api-k7m2xp"},
		{name: "app and source", appID: "billing-api-k7m2xp", sourceID: "stripe"},
		{name: "source without app", sourceID: "stripe", wantErr: true},
		{name: "invalid app", appID: "app_123", wantErr: true},
		{name: "invalid source", appID: "billing-api-k7m2xp", sourceID: "Stripe!", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListenSelectors(tt.appID, tt.sourceID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateListenSelectors() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
