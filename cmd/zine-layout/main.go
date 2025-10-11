package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	apicmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api"
	imagelayoutcmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/imagelayout"
	rendercmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/render"
	servecmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/serve"
	workflowcmd "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow"
	zldoc "github.com/go-go-golems/zine-layout/pkg/doc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// docs are embedded by pkg/doc/embed.go

var rootCmd = &cobra.Command{
	Use:   "zine-layout",
	Short: "Zine page layout engine",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := logging.InitLoggerFromViper()
		cobra.CheckErr(err)
	},
}

func initRoot() (*help.HelpSystem, error) {
	if err := logging.AddLoggingLayerToRootCommand(rootCmd, rootCmd.Use); err != nil {
		return nil, err
	}
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return nil, err
	}
	if err := logging.InitLoggerFromViper(); err != nil {
		return nil, err
	}
	hs := help.NewHelpSystem()
	// Load embedded docs into help system
	if err := hs.LoadSectionsFromFS(zldoc.DocsFS, "topics"); err != nil {
		return nil, err
	}
	help_cmd.SetupCobraRootCommand(hs, rootCmd)
	return hs, nil
}

func main() {
	hs, err := initRoot()
	cobra.CheckErr(err)
	_ = hs

	renderCmd, err := rendercmd.NewCommand()
	cobra.CheckErr(err)
	cobraRenderCmd, err := cli.BuildCobraCommandFromCommand(
		renderCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraRenderCmd)

	serveCmd, err := servecmd.NewCommand()
	cobra.CheckErr(err)
	cobraServeCmd, err := cli.BuildCobraCommandFromCommand(
		serveCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraServeCmd)

	// api group commands
	apiCmd, err := apicmd.NewCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(apiCmd)

	imagelayoutCmd, err := imagelayoutcmd.NewCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(imagelayoutCmd)

	workflowCmd, err := workflowcmd.NewCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(workflowCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}
