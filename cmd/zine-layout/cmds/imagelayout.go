package cmds

import (
	imagelayoutcmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/imagelayout"
	"github.com/spf13/cobra"
)

// NewImageLayoutCommand creates the imagelayout command group.
func NewImageLayoutCommand() (*cobra.Command, error) {
	return imagelayoutcmd.NewCommand(), nil
}
