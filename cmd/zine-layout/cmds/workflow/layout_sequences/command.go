package layoutsequencescmd

import "github.com/spf13/cobra"

// NewCommand builds the `workflow layout-sequences` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "layout-sequences",
		Short: "Manage layout sequences (laid-out image ordering) serverlessly",
	}

	factories := []func() (*cobra.Command, error){
		newLayoutSequencesListCommand,
		newLayoutSequencesGetCommand,
		newLayoutSequencesCreateCommand,
		newLayoutSequencesUpdateCommand,
		newLayoutSequencesDeleteCommand,
		newLayoutSequencesAddItemCommand,
		newLayoutSequencesDeleteItemCommand,
		newLayoutSequencesReorderCommand,
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
