package imagelayouttemplates

import "github.com/spf13/cobra"

// NewCommand returns the verb group command for image layout templates.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "image-layout-templates",
		Short: "Manage image layout templates",
	}

	factories := []func() (*cobra.Command, error){
		NewListCommand,
		NewGetCommand,
		NewCreateCommand,
		NewUpdateCommand,
		NewDeleteCommand,
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
