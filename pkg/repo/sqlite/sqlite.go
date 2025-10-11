package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

var _ = func() bool {
	rand.Seed(time.Now().UnixNano())
	return true
}()

// RunMigrations ensures the SQLite schema is created.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// NewRepositories constructs repository implementations backed by SQLite.
func NewRepositories(db *sql.DB) (*repo.Repositories, error) {
	if err := enablePragmas(db); err != nil {
		return nil, err
	}
	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	projects := &projectRepo{db: db}
	assets := &assetRepo{db: db}
	imageSequences := &imageSequenceRepo{db: db}
	layoutTemplates := &imageLayoutTemplateRepo{db: db}
	laidOutImages := &laidOutImageRepo{db: db}
	layoutSequences := &layoutSequenceRepo{db: db}

	return &repo.Repositories{
		Projects:             projects,
		Assets:               assets,
		ImageSequences:       imageSequences,
		ImageLayoutTemplates: layoutTemplates,
		LaidOutImages:        laidOutImages,
		LayoutSequences:      layoutSequences,
	}, nil
}

func enablePragmas(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	return nil
}

func scanNullableString(src sql.NullString) *string {
	if !src.Valid {
		return nil
	}
	s := src.String
	return &s
}

func nullStringPtr(v *string) sql.NullString {
	if v == nil || *v == "" {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: *v}
}

func toUnix(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UTC().Unix()
}

func fromUnix(v int64) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(v, 0).UTC()
}

func errNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	return err
}

func nullString(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: value}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// generateID creates a unique ID with timestamp and random suffix.
func generateID(prefix string) string {
	ts := time.Now().UTC().Format("20060102T150405Z")
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s-%s", prefix, ts, string(b))
}
