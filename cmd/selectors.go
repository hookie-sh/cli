package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/hookie-sh/cli/internal/publicid"
	"github.com/hookie-sh/cli/internal/relay"
)

func validateListenSelectors(appID, sourceID string) error {
	if sourceID != "" && appID == "" {
		return fmt.Errorf("--source-id requires --app-id (application public id)")
	}
	if appID != "" && !publicid.ValidAppPublicID(appID) {
		return fmt.Errorf("app_id must be a valid application public id (for example billing-api-k7m2xp)")
	}
	if sourceID != "" && !publicid.ValidSourceSlug(sourceID) {
		return fmt.Errorf("source_id must be a valid source slug (for example stripe)")
	}
	return nil
}

func promptListenTarget(ctx context.Context, client *relay.Client, orgID string) (string, string, error) {
	applications, err := client.ListApplications(ctx, orgID)
	if err != nil {
		return "", "", fmt.Errorf("failed to list applications: %w", relay.WrapRelayErr(err))
	}
	if len(applications) == 0 {
		return "", "", fmt.Errorf("no applications found. Create an application at https://app.hookie.sh first")
	}

	var selectedAppID string
	appNames := make(map[string]string, len(applications))
	appOptions := make([]huh.Option[string], 0, len(applications))
	for _, app := range applications {
		publicID := app.PublicId
		if publicID == "" {
			publicID = app.Id
		}
		name := app.Name
		if name == "" {
			name = publicID
		}
		appNames[publicID] = name
		appOptions = append(appOptions, huh.NewOption(
			fmt.Sprintf("%s (%s, %d sources)", name, publicID, app.SourceCount),
			publicID,
		))
	}

	appForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an application").
				Description("Choose the application to listen to").
				Options(appOptions...).
				Value(&selectedAppID),
		),
	)
	if err := appForm.RunWithContext(ctx); err != nil {
		return "", "", fmt.Errorf("failed to select application: %w", err)
	}
	if selectedAppID == "" {
		return "", "", fmt.Errorf("no application selected")
	}

	sources, err := client.ListSources(ctx, selectedAppID)
	if err != nil {
		return "", "", fmt.Errorf("failed to list sources: %w", relay.WrapRelayErr(err))
	}
	if len(sources) == 0 {
		return selectedAppID, "", nil
	}

	var selectedSourceID string
	sourceOptions := make([]huh.Option[string], 0, len(sources)+1)
	sourceOptions = append(sourceOptions, huh.NewOption("All sources", ""))
	for _, source := range sources {
		slug := source.Slug
		if slug == "" {
			slug = source.Id
		}
		name := source.Name
		if name == "" {
			name = slug
		}
		sourceOptions = append(sourceOptions, huh.NewOption(
			fmt.Sprintf("%s (%s)", name, slug),
			slug,
		))
	}

	sourceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Select a source from %s", appNames[selectedAppID])).
				Description("Choose one source, or listen to all sources").
				Options(sourceOptions...).
				Value(&selectedSourceID),
		),
	)
	if err := sourceForm.RunWithContext(ctx); err != nil {
		return "", "", fmt.Errorf("failed to select source: %w", err)
	}

	return selectedAppID, selectedSourceID, nil
}
