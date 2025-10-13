package imagesequencescmd

import "github.com/spf13/cobra"

// NewCommand constructs the `workflow image-sequences` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "image-sequences",
		Short: "Manage image sequences directly via the repositories",
	}

	factories := []func() (*cobra.Command, error){
		newImageSequencesListCommand,
		newImageSequencesGetCommand,
		newImageSequencesCreateCommand,
		newImageSequencesUpdateCommand,
		newImageSequencesDeleteCommand,
		newImageSequencesAddItemCommand,
		newImageSequencesDeleteItemCommand,
		newImageSequencesReorderCommand,
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
