package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type laidOutPageRepo struct {
	db *sql.DB
}

func (r *laidOutPageRepo) Create(page *repo.LaidOutPage) error {
	if page == nil {
		return fmt.Errorf("laid-out page is nil")
	}
	if page.ID == "" {
		page.ID = generateID("lpg")
	}
	now := time.Now().UTC()
	if page.CreatedAt.IsZero() {
		page.CreatedAt = now
	}
	if page.UpdatedAt.IsZero() {
		page.UpdatedAt = page.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO laid_out_pages (id, project_id, page_template_id, laid_out_image_id, result_json, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		page.ID,
		page.ProjectID,
		page.PageTemplateID,
		page.LaidOutImageID,
		nullStringPtr(page.ResultJSON),
		toUnix(page.CreatedAt),
		toUnix(page.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert laid-out page: %w", err)
	}
	return nil
}

func (r *laidOutPageRepo) Update(page *repo.LaidOutPage) error {
	if page == nil {
		return fmt.Errorf("laid-out page is nil")
	}
	if page.UpdatedAt.IsZero() {
		page.UpdatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(`UPDATE laid_out_pages
        SET page_template_id = ?, laid_out_image_id = ?, result_json = ?, updated_at = ?
        WHERE id = ?`,
		page.PageTemplateID,
		page.LaidOutImageID,
		nullStringPtr(page.ResultJSON),
		toUnix(page.UpdatedAt),
		page.ID,
	)
	if err != nil {
		return fmt.Errorf("update laid-out page: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *laidOutPageRepo) Get(id string) (*repo.LaidOutPage, error) {
	row := r.db.QueryRow(`SELECT id, project_id, page_template_id, laid_out_image_id, result_json, created_at, updated_at
        FROM laid_out_pages WHERE id = ?`, id)
	var (
		page    repo.LaidOutPage
		result  sql.NullString
		created int64
		updated int64
	)
	if err := row.Scan(&page.ID, &page.ProjectID, &page.PageTemplateID, &page.LaidOutImageID, &result, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	if result.Valid {
		page.ResultJSON = &result.String
	}
	page.CreatedAt = fromUnix(created)
	page.UpdatedAt = fromUnix(updated)
	return &page, nil
}

func (r *laidOutPageRepo) ListByProject(projectID string) ([]*repo.LaidOutPage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, page_template_id, laid_out_image_id, result_json, created_at, updated_at
        FROM laid_out_pages WHERE project_id = ?
        ORDER BY updated_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list laid-out pages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var pages []*repo.LaidOutPage
	for rows.Next() {
		var (
			page    repo.LaidOutPage
			result  sql.NullString
			created int64
			updated int64
		)
		if err := rows.Scan(&page.ID, &page.ProjectID, &page.PageTemplateID, &page.LaidOutImageID, &result, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan laid-out page: %w", err)
		}
		if result.Valid {
			page.ResultJSON = &result.String
		}
		page.CreatedAt = fromUnix(created)
		page.UpdatedAt = fromUnix(updated)
		pages = append(pages, &page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate laid-out pages: %w", err)
	}
	return pages, nil
}

func (r *laidOutPageRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM laid_out_pages WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete laid-out page: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
