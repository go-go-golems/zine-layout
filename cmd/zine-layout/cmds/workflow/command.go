package workflowcmd

import (
	assetscmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/assets"
	imagelayouttemplatescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/image_layout_templates"
	imagesequencescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/image_sequences"
	laidoutimagescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/laid_out_images"
	laidoutpagescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/laid_out_pages"
	layoutsequencescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/layout_sequences"
	pagetemplatescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/page_templates"
	projectscmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/projects"
	zinescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/zines"
	"github.com/spf13/cobra"
)

// NewCommand builds the `workflow` command group that operates directly on repositories.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "workflow",
		Short: "Low-level workflow helpers (direct repository access)",
	}

	pageTemplatesCmd, err := pagetemplatescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(pageTemplatesCmd)

	projectsCmd, err := projectscmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(projectsCmd)

	assetsCmd, err := assetscmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(assetsCmd)

	imageLayoutTemplatesCmd, err := imagelayouttemplatescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(imageLayoutTemplatesCmd)

	laidOutImagesCmd, err := laidoutimagescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(laidOutImagesCmd)

	imageSequencesCmd, err := imagesequencescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(imageSequencesCmd)

	layoutSequencesCmd, err := layoutsequencescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(layoutSequencesCmd)

	laidOutPagesCmd, err := laidoutpagescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(laidOutPagesCmd)

	zinesCmd, err := zinescmd.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(zinesCmd)

	return root, nil
}
