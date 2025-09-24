package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type spreadRepo struct {
	db *sql.DB
}

func (r *spreadRepo) Upsert(spread *repo.Spread) error {
	if spread == nil {
		return fmt.Errorf("spread is nil")
	}
	if spread.CreatedAt.IsZero() {
		spread.CreatedAt = time.Now().UTC()
	}
	if spread.UpdatedAt.IsZero() {
		spread.UpdatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO spreads (
        project_id, spread_number, left_page_number, right_page_number, settings_json, result_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(project_id, spread_number) DO UPDATE SET
        left_page_number = excluded.left_page_number,
        right_page_number = excluded.right_page_number,
        settings_json = excluded.settings_json,
        result_json = excluded.result_json,
        updated_at = excluded.updated_at`,
		spread.ProjectID,
		spread.SpreadNumber,
		nullIntPtr(spread.LeftPageNumber),
		nullIntPtr(spread.RightPageNumber),
		spread.SettingsJSON,
		nullStringPtr(spread.ResultJSON),
		toUnix(spread.CreatedAt),
		toUnix(spread.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert spread: %w", err)
	}
	return nil
}

func (r *spreadRepo) GetByNumber(projectID string, spreadNumber int) (*repo.Spread, error) {
	row := r.db.QueryRow(`SELECT project_id, spread_number, left_page_number, right_page_number, settings_json, result_json, created_at, updated_at
        FROM spreads WHERE project_id = ? AND spread_number = ?`, projectID, spreadNumber)
	var (
		spread           repo.Spread
		left             sql.NullInt64
		right            sql.NullInt64
		result           sql.NullString
		created, updated int64
	)
	if err := row.Scan(&spread.ProjectID, &spread.SpreadNumber, &left, &right, &spread.SettingsJSON, &result, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	spread.LeftPageNumber = scanNullableInt(left)
	spread.RightPageNumber = scanNullableInt(right)
	if result.Valid {
		val := result.String
		spread.ResultJSON = &val
	}
	spread.CreatedAt = fromUnix(created)
	spread.UpdatedAt = fromUnix(updated)
	return &spread, nil
}

func (r *spreadRepo) List(projectID string) ([]*repo.Spread, error) {
	rows, err := r.db.Query(`SELECT project_id, spread_number, left_page_number, right_page_number, settings_json, result_json, created_at, updated_at
        FROM spreads WHERE project_id = ? ORDER BY spread_number ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list spreads: %w", err)
	}
	defer rows.Close()

	var spreads []*repo.Spread
	for rows.Next() {
		var (
			spread           repo.Spread
			left             sql.NullInt64
			right            sql.NullInt64
			result           sql.NullString
			created, updated int64
		)
		if err := rows.Scan(&spread.ProjectID, &spread.SpreadNumber, &left, &right, &spread.SettingsJSON, &result, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan spread: %w", err)
		}
		spread.LeftPageNumber = scanNullableInt(left)
		spread.RightPageNumber = scanNullableInt(right)
		if result.Valid {
			val := result.String
			spread.ResultJSON = &val
		}
		spread.CreatedAt = fromUnix(created)
		spread.UpdatedAt = fromUnix(updated)
		spreads = append(spreads, &spread)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate spreads: %w", err)
	}
	return spreads, nil
}

func (r *spreadRepo) Delete(projectID string, spreadNumber int) error {
	res, err := r.db.Exec(`DELETE FROM spreads WHERE project_id = ? AND spread_number = ?`, projectID, spreadNumber)
	if err != nil {
		return fmt.Errorf("delete spread: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
