package laidoutimagescmd

import "github.com/spf13/cobra"

// NewCommand builds the `workflow laid-out-images` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "laid-out-images",
		Short: "Compute and manage laid-out images without the API server",
	}

	factories := []func() (*cobra.Command, error){
		newLaidOutImagesListCommand,
		newLaidOutImagesGetCommand,
		newLaidOutImagesCreateCommand,
		newLaidOutImagesUpdateCommand,
		newLaidOutImagesDeleteCommand,
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
