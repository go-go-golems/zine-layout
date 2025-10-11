package servecmd

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"

	servepkg "github.com/go-go-golems/zine-layout/pkg/serve"
)

type Command struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*Command)(nil)

func NewCommand() (*Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &Command{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Serve the Zine Layout web UI and API"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithDefault("./cmd/zine-layout/dist"), parameters.WithHelp("Path to built web assets (dist)")),
				parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to server data (projects, presets, uploads)")),
				parameters.NewParameterDefinition("addr", parameters.ParameterTypeString, parameters.WithDefault(":8088"), parameters.WithHelp("Listen address")),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

type ServeSettings struct {
	Root     string `glazed.parameter:"root"`
	DataRoot string `glazed.parameter:"data-root"`
	Addr     string `glazed.parameter:"addr"`
}

func (c *Command) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &ServeSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	srv := servepkg.New(servepkg.Settings{
		Root:     settings.Root,
		DataRoot: settings.DataRoot,
		Addr:     settings.Addr,
	})

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(runCtx)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	waitForServer := func() error {
		err := <-errCh
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	}

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig)
		cancel()
		return waitForServer()
	case <-ctx.Done():
		cancel()
		return waitForServer()
	}
}
