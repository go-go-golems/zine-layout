package laidoutpagescmd

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "laid-out-pages",
		Short: "Manage laid-out pages",
	}

	factories := []func() (*cobra.Command, error){
		newLaidOutPagesCreateCommand,
		newLaidOutPagesListCommand,
		newLaidOutPagesGetCommand,
		newLaidOutPagesUpdateImageCommand,
		newLaidOutPagesDeleteCommand,
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
