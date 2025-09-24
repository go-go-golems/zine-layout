package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type pageRepo struct {
	db *sql.DB
}

func (r *pageRepo) Upsert(page *repo.Page) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	if page.CreatedAt.IsZero() {
		page.CreatedAt = time.Now().UTC()
	}
	if page.UpdatedAt.IsZero() {
		page.UpdatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO pages (
        project_id, page_number, asset_id, settings_json, result_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(project_id, page_number) DO UPDATE SET
        asset_id = excluded.asset_id,
        settings_json = excluded.settings_json,
        result_json = excluded.result_json,
        updated_at = excluded.updated_at`,
		page.ProjectID,
		page.PageNumber,
		nullStringPtr(page.AssetID),
		page.SettingsJSON,
		page.ResultJSON,
		toUnix(page.CreatedAt),
		toUnix(page.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert page: %w", err)
	}
	return nil
}

func (r *pageRepo) GetByNumber(projectID string, pageNumber int) (*repo.Page, error) {
	row := r.db.QueryRow(`SELECT project_id, page_number, asset_id, settings_json, result_json, created_at, updated_at
        FROM pages WHERE project_id = ? AND page_number = ?`, projectID, pageNumber)
	var (
		page             repo.Page
		created, updated int64
		asset            sql.NullString
		result           sql.NullString
	)
	if err := row.Scan(&page.ProjectID, &page.PageNumber, &asset, &page.SettingsJSON, &result, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	page.AssetID = scanNullableString(asset)
	if result.Valid {
		page.ResultJSON = &result.String
	}
	page.CreatedAt = fromUnix(created)
	page.UpdatedAt = fromUnix(updated)
	return &page, nil
}

func (r *pageRepo) List(projectID string) ([]*repo.Page, error) {
	rows, err := r.db.Query(`SELECT project_id, page_number, asset_id, settings_json, result_json, created_at, updated_at
        FROM pages WHERE project_id = ? ORDER BY page_number ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}
	defer rows.Close()

	var pages []*repo.Page
	for rows.Next() {
		var (
			page             repo.Page
			created, updated int64
			asset            sql.NullString
			result           sql.NullString
		)
		if err := rows.Scan(&page.ProjectID, &page.PageNumber, &asset, &page.SettingsJSON, &result, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan page: %w", err)
		}
		page.AssetID = scanNullableString(asset)
		if result.Valid {
			val := result.String
			page.ResultJSON = &val
		}
		page.CreatedAt = fromUnix(created)
		page.UpdatedAt = fromUnix(updated)
		pages = append(pages, &page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pages: %w", err)
	}
	return pages, nil
}

func (r *pageRepo) Delete(projectID string, pageNumber int) error {
	res, err := r.db.Exec(`DELETE FROM pages WHERE project_id = ? AND page_number = ?`, projectID, pageNumber)
	if err != nil {
		return fmt.Errorf("delete page: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
