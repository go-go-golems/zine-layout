package laidoutimages

import "github.com/spf13/cobra"

// NewCommand returns the verb group command for laid-out images.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "laid-out-images",
		Short: "Manage laid-out images",
	}

	factories := []func() (*cobra.Command, error){
		NewListCommand,
		NewGetCommand,
		NewCreateCommand,
		NewUpdateCommand,
		NewDeleteCommand,
		NewPreviewCommand,
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
