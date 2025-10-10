package imagesequences

import "github.com/spf13/cobra"

// NewCommand returns the verb group command for image sequences.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{Use: "image-sequences", Short: "Manage image sequences"}

	factories := []func() (*cobra.Command, error){
		NewListCommand,
		NewGetCommand,
		NewCreateCommand,
		NewUpdateCommand,
		NewDeleteCommand,
		NewAddItemCommand,
		NewReorderCommand,
		NewDeleteItemCommand,
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
