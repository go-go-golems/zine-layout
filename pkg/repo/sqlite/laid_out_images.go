package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type laidOutImageRepo struct {
	db *sql.DB
}

func (r *laidOutImageRepo) Create(image *repo.LaidOutImage) error {
	if image == nil {
		return fmt.Errorf("laid-out image is nil")
	}
	if image.ID == "" {
		image.ID = generateID("loi")
	}
	now := time.Now().UTC()
	if image.CreatedAt.IsZero() {
		image.CreatedAt = now
	}
	if image.UpdatedAt.IsZero() {
		image.UpdatedAt = image.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO laid_out_images
        (id, project_id, asset_id, template_id, overrides_json, result_json, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		image.ID,
		image.ProjectID,
		image.AssetID,
		image.TemplateID,
		nullStringPtr(image.OverridesJSON),
		image.ResultJSON,
		toUnix(image.CreatedAt),
		toUnix(image.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert laid-out image: %w", err)
	}
	return nil
}

func (r *laidOutImageRepo) Update(image *repo.LaidOutImage) error {
	if image == nil {
		return fmt.Errorf("laid-out image is nil")
	}
	if image.UpdatedAt.IsZero() {
		image.UpdatedAt = time.Now().UTC()
	}
	res, err := r.db.Exec(`UPDATE laid_out_images
        SET project_id = ?, asset_id = ?, template_id = ?, overrides_json = ?, result_json = ?, updated_at = ?
        WHERE id = ?`,
		image.ProjectID,
		image.AssetID,
		image.TemplateID,
		nullStringPtr(image.OverridesJSON),
		image.ResultJSON,
		toUnix(image.UpdatedAt),
		image.ID,
	)
	if err != nil {
		return fmt.Errorf("update laid-out image: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *laidOutImageRepo) Get(id string) (*repo.LaidOutImage, error) {
	row := r.db.QueryRow(`SELECT id, project_id, asset_id, template_id, overrides_json, result_json, created_at, updated_at
        FROM laid_out_images WHERE id = ?`, id)
	var (
		image     repo.LaidOutImage
		overrides sql.NullString
		created   int64
		updated   int64
	)
	if err := row.Scan(&image.ID, &image.ProjectID, &image.AssetID, &image.TemplateID, &overrides, &image.ResultJSON, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	image.OverridesJSON = scanNullableString(overrides)
	image.CreatedAt = fromUnix(created)
	image.UpdatedAt = fromUnix(updated)
	return &image, nil
}

func (r *laidOutImageRepo) ListByProject(projectID string) ([]*repo.LaidOutImage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, asset_id, template_id, overrides_json, result_json, created_at, updated_at
        FROM laid_out_images WHERE project_id = ?
        ORDER BY updated_at DESC, id ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list laid-out images by project: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var images []*repo.LaidOutImage
	for rows.Next() {
		var (
			image     repo.LaidOutImage
			overrides sql.NullString
			created   int64
			updated   int64
		)
		if err := rows.Scan(&image.ID, &image.ProjectID, &image.AssetID, &image.TemplateID, &overrides, &image.ResultJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan laid-out image: %w", err)
		}
		image.OverridesJSON = scanNullableString(overrides)
		image.CreatedAt = fromUnix(created)
		image.UpdatedAt = fromUnix(updated)
		images = append(images, &image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate laid-out images: %w", err)
	}
	return images, nil
}

func (r *laidOutImageRepo) ListByAsset(assetID string) ([]*repo.LaidOutImage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, asset_id, template_id, overrides_json, result_json, created_at, updated_at
        FROM laid_out_images WHERE asset_id = ?
        ORDER BY updated_at DESC, id ASC`, assetID)
	if err != nil {
		return nil, fmt.Errorf("list laid-out images by asset: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var images []*repo.LaidOutImage
	for rows.Next() {
		var (
			image     repo.LaidOutImage
			overrides sql.NullString
			created   int64
			updated   int64
		)
		if err := rows.Scan(&image.ID, &image.ProjectID, &image.AssetID, &image.TemplateID, &overrides, &image.ResultJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan laid-out image: %w", err)
		}
		image.OverridesJSON = scanNullableString(overrides)
		image.CreatedAt = fromUnix(created)
		image.UpdatedAt = fromUnix(updated)
		images = append(images, &image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate laid-out images: %w", err)
	}
	return images, nil
}

func (r *laidOutImageRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM laid_out_images WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete laid-out image: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
