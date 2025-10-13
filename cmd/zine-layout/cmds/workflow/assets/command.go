package assetscmd

import "github.com/spf13/cobra"

// NewCommand constructs the `workflow assets` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "assets",
		Short: "Manage project assets without the HTTP API",
	}

	factories := []func() (*cobra.Command, error){
		newAssetsListCommand,
		newAssetsCreateCommand,
	}

	for _, factory := range factories {
		cmd, err := factory()
		if err != nil {
			return nil, err
		}
		root.AddCommand(cmd)
	}

	return root, nil
}
