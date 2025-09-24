package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type projectRepo struct {
	db *sql.DB
}

func (r *projectRepo) Create(project *repo.Project) error {
	if project == nil {
		return fmt.Errorf("project is nil")
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now().UTC()
	}
	if project.UpdatedAt.IsZero() {
		project.UpdatedAt = project.CreatedAt
	}
	_, err := r.db.Exec(`INSERT INTO projects (id, name, created_at, updated_at, preset_id, cover_asset_id)
VALUES (?, ?, ?, ?, ?, ?)`,
		project.ID,
		project.Name,
		toUnix(project.CreatedAt),
		toUnix(project.UpdatedAt),
		nullStringPtr(project.PresetID),
		nullStringPtr(project.CoverAssetID),
	)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *projectRepo) Update(project *repo.Project) error {
	if project == nil {
		return fmt.Errorf("project is nil")
	}
	if project.UpdatedAt.IsZero() {
		project.UpdatedAt = time.Now().UTC()
	}
	res, err := r.db.Exec(`UPDATE projects SET name = ?, updated_at = ?, preset_id = ?, cover_asset_id = ? WHERE id = ?`,
		project.Name,
		toUnix(project.UpdatedAt),
		nullStringPtr(project.PresetID),
		nullStringPtr(project.CoverAssetID),
		project.ID,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *projectRepo) Get(id string) (*repo.Project, error) {
	row := r.db.QueryRow(`SELECT id, name, created_at, updated_at, preset_id, cover_asset_id FROM projects WHERE id = ?`, id)
	var (
		project          repo.Project
		created, updated int64
		preset           sql.NullString
		cover            sql.NullString
	)
	if err := row.Scan(&project.ID, &project.Name, &created, &updated, &preset, &cover); err != nil {
		return nil, errNotFound(err)
	}
	project.CreatedAt = fromUnix(created)
	project.UpdatedAt = fromUnix(updated)
	project.PresetID = scanNullableString(preset)
	project.CoverAssetID = scanNullableString(cover)
	return &project, nil
}

func (r *projectRepo) List() ([]*repo.Project, error) {
	rows, err := r.db.Query(`SELECT id, name, created_at, updated_at, preset_id, cover_asset_id FROM projects ORDER BY updated_at DESC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []*repo.Project
	for rows.Next() {
		var (
			project          repo.Project
			created, updated int64
			preset           sql.NullString
			cover            sql.NullString
		)
		if err := rows.Scan(&project.ID, &project.Name, &created, &updated, &preset, &cover); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		project.CreatedAt = fromUnix(created)
		project.UpdatedAt = fromUnix(updated)
		project.PresetID = scanNullableString(preset)
		project.CoverAssetID = scanNullableString(cover)
		projects = append(projects, &project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (r *projectRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
