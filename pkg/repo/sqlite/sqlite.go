package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

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
	pages := &pageRepo{db: db}
	spreads := &spreadRepo{db: db}

	return &repo.Repositories{
		Projects: projects,
		Assets:   assets,
		Pages:    pages,
		Spreads:  spreads,
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

func scanNullableInt(src sql.NullInt64) *int {
	if !src.Valid {
		return nil
	}
	v := int(src.Int64)
	return &v
}

func nullIntPtr(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Valid: true, Int64: int64(*v)}
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
