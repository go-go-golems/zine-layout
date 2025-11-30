package pagelayoutcmd

import "github.com/spf13/cobra"

// NewCommand builds the pagelayout verb-group root and wires subcommands.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "pagelayout",
		Short: "Page layout helpers",
	}

	computeCmd, err := NewComputeCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(computeCmd)

	renderCmd, err := NewRenderCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(renderCmd)

	return root, nil
}

