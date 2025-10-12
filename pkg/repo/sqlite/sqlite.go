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
	if err := prepareSchema(db); err != nil {
		return err
	}
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
	pageTemplates := &pageTemplateRepo{db: db}
	laidOutPages := &laidOutPageRepo{db: db}
	zines := &zineRepo{db: db}

	return &repo.Repositories{
		Projects:             projects,
		Assets:               assets,
		ImageSequences:       imageSequences,
		ImageLayoutTemplates: layoutTemplates,
		LaidOutImages:        laidOutImages,
		LayoutSequences:      layoutSequences,
		PageTemplates:        pageTemplates,
		LaidOutPages:         laidOutPages,
		Zines:                zines,
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

func prepareSchema(db *sql.DB) error {
	// Legacy Phase 2 table lacked laid_out_image_id; drop and recreate if necessary.
	hasPagesTable, err := hasTable(db, "laid_out_pages")
	if err != nil {
		return err
	}
	if hasPagesTable {
		hasColumn, err := tableHasColumn(db, "laid_out_pages", "laid_out_image_id")
		if err != nil {
			return err
		}
		if !hasColumn {
			if _, err := db.Exec(`DROP TABLE IF EXISTS zine_pages;`); err != nil {
				return fmt.Errorf("drop legacy zine_pages: %w", err)
			}
			if _, err := db.Exec(`DROP TABLE IF EXISTS laid_out_pages;`); err != nil {
				return fmt.Errorf("drop legacy laid_out_pages: %w", err)
			}
		}
	}
	// Remove legacy inputs table if it still exists.
	if _, err := db.Exec(`DROP TABLE IF EXISTS laid_out_page_inputs;`); err != nil {
		return fmt.Errorf("drop legacy laid_out_page_inputs: %w", err)
	}
	return nil
}

func hasTable(db *sql.DB, table string) (bool, error) {
	row := db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type='table' AND name=?`, table)
	var dummy int
	err := row.Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query sqlite_master for %s: %w", table, err)
	}
	return true, nil
}

func tableHasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, fmt.Errorf("pragma table_info(%s): %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid        int
			name       string
			ctype      string
			notNull    int
			dfltValue  any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &primaryKey); err != nil {
			return false, fmt.Errorf("scan table_info(%s): %w", table, err)
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate table_info(%s): %w", table, err)
	}
	return false, nil
}
