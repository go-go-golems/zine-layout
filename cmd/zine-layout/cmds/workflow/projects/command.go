package projectscmd

import "github.com/spf13/cobra"

// NewCommand wires the `workflow projects` command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects directly against the repositories",
	}

	factories := []func() (*cobra.Command, error){
		newProjectsListCommand,
		newProjectsGetCommand,
		newProjectsCreateCommand,
		newProjectsUpdateCommand,
		newProjectsDeleteCommand,
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
