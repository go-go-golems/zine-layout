package pagetemplatescmd

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "page-templates",
		Short: "Direct repository access to page templates",
	}

	factories := []func() (*cobra.Command, error){
		newPageTemplatesListCommand,
		newPageTemplatesGetCommand,
		newPageTemplatesCreateCommand,
		newPageTemplatesDeleteCommand,
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
