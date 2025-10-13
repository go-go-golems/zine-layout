package workflowshared

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/zine-layout/pkg/repo"
	sqliterepo "github.com/go-go-golems/zine-layout/pkg/repo/sqlite"
	"github.com/spf13/cobra"
)

func BuildCommand(cmd cmds.GlazeCommand) (*cobra.Command, error) {
	return cli.BuildCobraCommand(cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
}

func OpenRepositories(dataRoot string) (*repo.Repositories, *sql.DB, error) {
	if dataRoot == "" {
		dataRoot = "./data"
	}
	if err := os.MkdirAll(dataRoot, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create data root: %w", err)
	}
	dbPath := filepath.Join(dataRoot, "zine-layout.db")
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite: %w", err)
	}
	repos, err := sqliterepo.NewRepositories(db)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return repos, db, nil
}

func Stringify(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ProjectsRoot returns the path where project data is stored for the provided data root.
func ProjectsRoot(dataRoot string) string {
	if dataRoot == "" {
		dataRoot = "./data"
	}
	return filepath.Join(dataRoot, "projects")
}
