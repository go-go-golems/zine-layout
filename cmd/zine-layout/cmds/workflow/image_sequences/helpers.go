package imagesequencescmd

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

func parseItemTokens(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		lower := strings.ToLower(token)
		if lower == "gap" || lower == "_" || lower == "-" {
			items = append(items, "")
			continue
		}
		items = append(items, token)
	}
	return items
}

func ensureAssetInProject(repos *repo.Repositories, projectID, assetID string) error {
	asset, err := repos.Assets.Get(assetID)
	if err != nil {
		return fmt.Errorf("lookup asset %s: %w", assetID, err)
	}
	if asset.ProjectID != projectID {
		return fmt.Errorf("asset %s does not belong to project %s", assetID, projectID)
	}
	return nil
}
