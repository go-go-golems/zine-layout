package workflowcmd

import (
	laidoutpagescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/laid_out_pages"
	pagetemplatescmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/page_templates"
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
