package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type zineRepo struct {
	db *sql.DB
}

func (r *zineRepo) Create(zine *repo.Zine) error {
	if zine == nil {
		return fmt.Errorf("zine is nil")
	}
	if zine.ID == "" {
		zine.ID = generateID("zne")
	}
	now := time.Now().UTC()
	if zine.CreatedAt.IsZero() {
		zine.CreatedAt = now
	}
	if zine.UpdatedAt.IsZero() {
		zine.UpdatedAt = zine.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO zines (id, project_id, name, description, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`,
		zine.ID,
		zine.ProjectID,
		zine.Name,
		zine.Description,
		toUnix(zine.CreatedAt),
		toUnix(zine.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert zine: %w", err)
	}
	return nil
}

func (r *zineRepo) Update(zine *repo.Zine) error {
	if zine == nil {
		return fmt.Errorf("zine is nil")
	}
	if zine.UpdatedAt.IsZero() {
		zine.UpdatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(`UPDATE zines
        SET name = ?, description = ?, updated_at = ?
        WHERE id = ?`,
		zine.Name,
		zine.Description,
		toUnix(zine.UpdatedAt),
		zine.ID,
	)
	if err != nil {
		return fmt.Errorf("update zine: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *zineRepo) Get(id string) (*repo.Zine, error) {
	row := r.db.QueryRow(`SELECT id, project_id, name, description, created_at, updated_at
        FROM zines WHERE id = ?`, id)
	var (
		zine    repo.Zine
		created int64
		updated int64
	)
	if err := row.Scan(&zine.ID, &zine.ProjectID, &zine.Name, &zine.Description, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	zine.CreatedAt = fromUnix(created)
	zine.UpdatedAt = fromUnix(updated)
	return &zine, nil
}

func (r *zineRepo) ListByProject(projectID string) ([]*repo.Zine, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, created_at, updated_at
        FROM zines WHERE project_id = ?
        ORDER BY updated_at DESC, name ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list zines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var zines []*repo.Zine
	for rows.Next() {
		var (
			zine    repo.Zine
			created int64
			updated int64
		)
		if err := rows.Scan(&zine.ID, &zine.ProjectID, &zine.Name, &zine.Description, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan zine: %w", err)
		}
		zine.CreatedAt = fromUnix(created)
		zine.UpdatedAt = fromUnix(updated)
		zines = append(zines, &zine)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate zines: %w", err)
	}
	return zines, nil
}

func (r *zineRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM zines WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete zine: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *zineRepo) SetPages(zineID string, pages []*repo.ZinePage) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin set zine pages tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM zine_pages WHERE zine_id = ?`, zineID); err != nil {
		return fmt.Errorf("delete existing zine pages: %w", err)
	}

	for _, page := range pages {
		if page == nil {
			continue
		}
		if _, err = tx.Exec(`INSERT INTO zine_pages (zine_id, position, laid_out_page_id)
            VALUES (?, ?, ?)`,
			zineID,
			page.Position,
			page.LaidOutPageID,
		); err != nil {
			return fmt.Errorf("insert zine page: %w", err)
		}
	}
	return nil
}

func (r *zineRepo) GetPages(zineID string) ([]*repo.ZinePage, error) {
	rows, err := r.db.Query(`SELECT zine_id, position, laid_out_page_id
        FROM zine_pages WHERE zine_id = ?
        ORDER BY position ASC`, zineID)
	if err != nil {
		return nil, fmt.Errorf("list zine pages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var pages []*repo.ZinePage
	for rows.Next() {
		var page repo.ZinePage
		if err := rows.Scan(&page.ZineID, &page.Position, &page.LaidOutPageID); err != nil {
			return nil, fmt.Errorf("scan zine page: %w", err)
		}
		pages = append(pages, &page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate zine pages: %w", err)
	}
	return pages, nil
}
