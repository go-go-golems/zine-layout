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
	if asset.ID == "" {
		asset.ID = generateID("ast")
	}
	if asset.UploadedAt.IsZero() {
		asset.UploadedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO assets (
        id, project_id, filename, rel_path, content_type, bytes, width, height, uploaded_at, metadata_json
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		asset.ID,
		asset.ProjectID,
		asset.Filename,
		asset.RelPath,
		asset.ContentType,
		asset.Bytes,
		asset.Width,
		asset.Height,
		toUnix(asset.UploadedAt),
		nullString(asset.MetadataJSON),
	)
	if err != nil {
		return fmt.Errorf("insert asset: %w", err)
	}
	return nil
}

func (r *assetRepo) Update(asset *repo.Asset) error {
	if asset == nil {
		return fmt.Errorf("asset is nil")
	}
	if asset.UploadedAt.IsZero() {
		asset.UploadedAt = time.Now().UTC()
	}
	res, err := r.db.Exec(`UPDATE assets
        SET filename = ?, rel_path = ?, content_type = ?, bytes = ?, width = ?, height = ?, uploaded_at = ?, metadata_json = ?
        WHERE id = ?`,
		asset.Filename,
		asset.RelPath,
		asset.ContentType,
		asset.Bytes,
		asset.Width,
		asset.Height,
		toUnix(asset.UploadedAt),
		nullString(asset.MetadataJSON),
		asset.ID,
	)
	if err != nil {
		return fmt.Errorf("update asset: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *assetRepo) Get(id string) (*repo.Asset, error) {
	row := r.db.QueryRow(`SELECT id, project_id, filename, rel_path, content_type, bytes, width, height, uploaded_at, metadata_json
        FROM assets WHERE id = ?`, id)
	var (
		asset    repo.Asset
		uploaded int64
		content  sql.NullString
		metadata sql.NullString
	)
	if err := row.Scan(&asset.ID, &asset.ProjectID, &asset.Filename, &asset.RelPath, &content, &asset.Bytes, &asset.Width, &asset.Height, &uploaded, &metadata); err != nil {
		return nil, errNotFound(err)
	}
	asset.ContentType = content.String
	asset.MetadataJSON = metadata.String
	asset.UploadedAt = fromUnix(uploaded)
	return &asset, nil
}

func (r *assetRepo) ListByProject(projectID string) ([]*repo.Asset, error) {
	rows, err := r.db.Query(`SELECT id, project_id, filename, rel_path, content_type, bytes, width, height, uploaded_at, metadata_json
        FROM assets WHERE project_id = ? ORDER BY uploaded_at ASC, id ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var assets []*repo.Asset
	for rows.Next() {
		var (
			asset    repo.Asset
			uploaded int64
			content  sql.NullString
			metadata sql.NullString
		)
		if err := rows.Scan(&asset.ID, &asset.ProjectID, &asset.Filename, &asset.RelPath, &content, &asset.Bytes, &asset.Width, &asset.Height, &uploaded, &metadata); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		asset.ContentType = content.String
		asset.MetadataJSON = metadata.String
		asset.UploadedAt = fromUnix(uploaded)
		assets = append(assets, &asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}
	return assets, nil
}

func (r *assetRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM assets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
