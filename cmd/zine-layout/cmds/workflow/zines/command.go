package zinescmd

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "zines",
		Short: "Manage zines",
	}

	factories := []func() (*cobra.Command, error){
		newZinesCreateCommand,
		newZinesListCommand,
		newZinesGetCommand,
		newZinesSetPagesCommand,
		newZinesDeleteCommand,
		newZinesExportCommand,
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
