package imagelayouttemplatescmd

import "github.com/spf13/cobra"

// NewCommand builds the `workflow image-layout-templates` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "image-layout-templates",
		Short: "Manage image layout templates without the API server",
	}

	factories := []func() (*cobra.Command, error){
		newTemplatesListCommand,
		newTemplatesGetCommand,
		newTemplatesCreateCommand,
		newTemplatesUpdateCommand,
		newTemplatesDeleteCommand,
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
