package imagelayoutcmd

import "github.com/spf13/cobra"

// NewCommand builds the imagelayout verb-group root and wires subcommands.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "imagelayout",
		Short: "Local image layout helpers",
	}

	computeCmd, err := NewComputeCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(computeCmd)

	root.AddCommand(newLayoutCommand())

	return root, nil
}
