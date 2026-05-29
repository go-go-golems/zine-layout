package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type imageLayoutTemplateRepo struct {
	db *sql.DB
}

func (r *imageLayoutTemplateRepo) Create(tpl *repo.ImageLayoutTemplate) error {
	if tpl == nil {
		return fmt.Errorf("template is nil")
	}
	if tpl.ID == "" {
		tpl.ID = generateID("tpl")
	}
	now := time.Now().UTC()
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = now
	}
	if tpl.UpdatedAt.IsZero() {
		tpl.UpdatedAt = tpl.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO image_layout_templates (id, project_id, name, description, settings_json, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tpl.ID,
		nullStringPtr(tpl.ProjectID),
		tpl.Name,
		tpl.Description,
		tpl.SettingsJSON,
		toUnix(tpl.CreatedAt),
		toUnix(tpl.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert image layout template: %w", err)
	}
	return nil
}

func (r *imageLayoutTemplateRepo) Update(tpl *repo.ImageLayoutTemplate) error {
	if tpl == nil {
		return fmt.Errorf("template is nil")
	}
	if tpl.UpdatedAt.IsZero() {
		tpl.UpdatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(`UPDATE image_layout_templates
        SET project_id = ?, name = ?, description = ?, settings_json = ?, updated_at = ?
        WHERE id = ?`,
		nullStringPtr(tpl.ProjectID),
		tpl.Name,
		tpl.Description,
		tpl.SettingsJSON,
		toUnix(tpl.UpdatedAt),
		tpl.ID,
	)
	if err != nil {
		return fmt.Errorf("update image layout template: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *imageLayoutTemplateRepo) Get(id string) (*repo.ImageLayoutTemplate, error) {
	row := r.db.QueryRow(`SELECT id, project_id, name, description, settings_json, created_at, updated_at
        FROM image_layout_templates WHERE id = ?`, id)
	var (
		tpl       repo.ImageLayoutTemplate
		projectID sql.NullString
		created   int64
		updated   int64
	)
	if err := row.Scan(&tpl.ID, &projectID, &tpl.Name, &tpl.Description, &tpl.SettingsJSON, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	tpl.ProjectID = scanNullableString(projectID)
	tpl.CreatedAt = fromUnix(created)
	tpl.UpdatedAt = fromUnix(updated)
	return &tpl, nil
}

func (r *imageLayoutTemplateRepo) ListGlobal() ([]*repo.ImageLayoutTemplate, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, settings_json, created_at, updated_at
        FROM image_layout_templates WHERE project_id IS NULL
        ORDER BY updated_at DESC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list global templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []*repo.ImageLayoutTemplate
	for rows.Next() {
		var (
			tpl       repo.ImageLayoutTemplate
			projectID sql.NullString
			created   int64
			updated   int64
		)
		if err := rows.Scan(&tpl.ID, &projectID, &tpl.Name, &tpl.Description, &tpl.SettingsJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan image layout template: %w", err)
		}
		tpl.ProjectID = scanNullableString(projectID)
		tpl.CreatedAt = fromUnix(created)
		tpl.UpdatedAt = fromUnix(updated)
		templates = append(templates, &tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate image layout templates: %w", err)
	}
	return templates, nil
}

func (r *imageLayoutTemplateRepo) ListByProject(projectID string) ([]*repo.ImageLayoutTemplate, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, settings_json, created_at, updated_at
        FROM image_layout_templates WHERE project_id = ?
        ORDER BY updated_at DESC, name ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []*repo.ImageLayoutTemplate
	for rows.Next() {
		var (
			tpl         repo.ImageLayoutTemplate
			projectNull sql.NullString
			created     int64
			updated     int64
		)
		if err := rows.Scan(&tpl.ID, &projectNull, &tpl.Name, &tpl.Description, &tpl.SettingsJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan image layout template: %w", err)
		}
		tpl.ProjectID = scanNullableString(projectNull)
		tpl.CreatedAt = fromUnix(created)
		tpl.UpdatedAt = fromUnix(updated)
		templates = append(templates, &tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate image layout templates: %w", err)
	}
	return templates, nil
}

func (r *imageLayoutTemplateRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM image_layout_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete image layout template: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
