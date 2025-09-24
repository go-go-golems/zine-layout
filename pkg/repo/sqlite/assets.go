package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type assetRepo struct {
	db *sql.DB
}

func (r *assetRepo) Create(asset *repo.Asset) error {
	if asset == nil {
		return fmt.Errorf("asset is nil")
	}
	if asset.CreatedAt.IsZero() {
		asset.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO assets (
        project_id, id, filename, rel_path, content_type, bytes, width, height, sort_index, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(project_id, id) DO UPDATE SET
        filename = excluded.filename,
        rel_path = excluded.rel_path,
        content_type = excluded.content_type,
        bytes = excluded.bytes,
        width = excluded.width,
        height = excluded.height,
        sort_index = excluded.sort_index,
        created_at = excluded.created_at`,
		asset.ProjectID,
		asset.ID,
		asset.Filename,
		asset.RelPath,
		asset.ContentType,
		asset.Bytes,
		asset.Width,
		asset.Height,
		asset.SortIndex,
		toUnix(asset.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert asset: %w", err)
	}
	return nil
}

func (r *assetRepo) Get(projectID, assetID string) (*repo.Asset, error) {
	row := r.db.QueryRow(`SELECT project_id, id, filename, rel_path, content_type, bytes, width, height, sort_index, created_at FROM assets WHERE project_id = ? AND id = ?`, projectID, assetID)
	var (
		asset   repo.Asset
		created int64
		content sql.NullString
	)
	if err := row.Scan(&asset.ProjectID, &asset.ID, &asset.Filename, &asset.RelPath, &content, &asset.Bytes, &asset.Width, &asset.Height, &asset.SortIndex, &created); err != nil {
		return nil, errNotFound(err)
	}
	asset.ContentType = content.String
	asset.CreatedAt = fromUnix(created)
	return &asset, nil
}

func (r *assetRepo) ListByProject(projectID string) ([]*repo.Asset, error) {
	rows, err := r.db.Query(`SELECT project_id, id, filename, rel_path, content_type, bytes, width, height, sort_index, created_at
        FROM assets WHERE project_id = ? ORDER BY sort_index ASC, created_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var assets []*repo.Asset
	for rows.Next() {
		var (
			asset   repo.Asset
			created int64
			content sql.NullString
		)
		if err := rows.Scan(&asset.ProjectID, &asset.ID, &asset.Filename, &asset.RelPath, &content, &asset.Bytes, &asset.Width, &asset.Height, &asset.SortIndex, &created); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		asset.ContentType = content.String
		asset.CreatedAt = fromUnix(created)
		assets = append(assets, &asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}
	return assets, nil
}

func (r *assetRepo) UpdateOrder(projectID string, orderedIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin asset order tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	for idx, id := range orderedIDs {
		if _, execErr := tx.Exec(`UPDATE assets SET sort_index = ? WHERE project_id = ? AND id = ?`, idx, projectID, id); execErr != nil {
			err = fmt.Errorf("update sort index: %w", execErr)
			return err
		}
	}
	return nil
}

func (r *assetRepo) Delete(projectID, assetID string) error {
	res, err := r.db.Exec(`DELETE FROM assets WHERE project_id = ? AND id = ?`, projectID, assetID)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
