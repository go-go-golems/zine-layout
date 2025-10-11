package workflowcmd

import (
	"github.com/spf13/cobra"
)

// NewCommand builds the `workflow` command group that operates directly on repositories.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "workflow",
		Short: "Low-level workflow helpers (direct repository access)",
	}

	pageTemplatesCmd, err := newPageTemplatesCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(pageTemplatesCmd)

	laidOutPagesCmd, err := newLaidOutPagesCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(laidOutPagesCmd)

	zinesCmd, err := newZinesCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(zinesCmd)

	return root, nil
}
