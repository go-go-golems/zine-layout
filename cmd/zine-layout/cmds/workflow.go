package cmds

import (
	"github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow"
	"github.com/spf13/cobra"
)

// NewWorkflowCommand returns the `workflow` command namespace.
func NewWorkflowCommand() (*cobra.Command, error) {
	return workflowcmd.NewCommand()
}
