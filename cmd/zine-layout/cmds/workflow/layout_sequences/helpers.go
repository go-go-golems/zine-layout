package layoutsequencescmd

import (
	"fmt"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

func ensureLaidOutImageInProject(repos *repo.Repositories, projectID, laidOutImageID string) error {
	image, err := repos.LaidOutImages.Get(laidOutImageID)
	if err != nil {
		return fmt.Errorf("lookup laid-out image %s: %w", laidOutImageID, err)
	}
	if image.ProjectID != projectID {
		return fmt.Errorf("laid-out image %s does not belong to project %s", laidOutImageID, projectID)
	}
	return nil
}
