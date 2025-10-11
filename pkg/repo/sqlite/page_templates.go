package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type pageTemplateRepo struct {
	db *sql.DB
}

func (r *pageTemplateRepo) Create(tpl *repo.PageTemplate) error {
	if tpl == nil {
		return fmt.Errorf("page template is nil")
	}
	if tpl.ID == "" {
		tpl.ID = generateID("ptpl")
	}
	now := time.Now().UTC()
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = now
	}
	if tpl.UpdatedAt.IsZero() {
		tpl.UpdatedAt = tpl.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO page_templates (id, project_id, name, description, template_json, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tpl.ID,
		nullStringPtr(tpl.ProjectID),
		tpl.Name,
		tpl.Description,
		tpl.TemplateJSON,
		toUnix(tpl.CreatedAt),
		toUnix(tpl.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert page template: %w", err)
	}
	return nil
}

func (r *pageTemplateRepo) Update(tpl *repo.PageTemplate) error {
	if tpl == nil {
		return fmt.Errorf("page template is nil")
	}
	if tpl.UpdatedAt.IsZero() {
		tpl.UpdatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(`UPDATE page_templates
        SET project_id = ?, name = ?, description = ?, template_json = ?, updated_at = ?
        WHERE id = ?`,
		nullStringPtr(tpl.ProjectID),
		tpl.Name,
		tpl.Description,
		tpl.TemplateJSON,
		toUnix(tpl.UpdatedAt),
		tpl.ID,
	)
	if err != nil {
		return fmt.Errorf("update page template: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *pageTemplateRepo) Get(id string) (*repo.PageTemplate, error) {
	row := r.db.QueryRow(`SELECT id, project_id, name, description, template_json, created_at, updated_at
        FROM page_templates WHERE id = ?`, id)
	var (
		tpl       repo.PageTemplate
		projectID sql.NullString
		created   int64
		updated   int64
	)
	if err := row.Scan(&tpl.ID, &projectID, &tpl.Name, &tpl.Description, &tpl.TemplateJSON, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	tpl.ProjectID = scanNullableString(projectID)
	tpl.CreatedAt = fromUnix(created)
	tpl.UpdatedAt = fromUnix(updated)
	return &tpl, nil
}

func (r *pageTemplateRepo) ListGlobal() ([]*repo.PageTemplate, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, template_json, created_at, updated_at
        FROM page_templates WHERE project_id IS NULL
        ORDER BY updated_at DESC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list global page templates: %w", err)
	}
	defer rows.Close()

	var templates []*repo.PageTemplate
	for rows.Next() {
		var (
			tpl       repo.PageTemplate
			projectID sql.NullString
			created   int64
			updated   int64
		)
		if err := rows.Scan(&tpl.ID, &projectID, &tpl.Name, &tpl.Description, &tpl.TemplateJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan page template: %w", err)
		}
		tpl.ProjectID = scanNullableString(projectID)
		tpl.CreatedAt = fromUnix(created)
		tpl.UpdatedAt = fromUnix(updated)
		templates = append(templates, &tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate page templates: %w", err)
	}
	return templates, nil
}

func (r *pageTemplateRepo) ListByProject(projectID string) ([]*repo.PageTemplate, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, template_json, created_at, updated_at
        FROM page_templates WHERE project_id = ?
        ORDER BY updated_at DESC, name ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project page templates: %w", err)
	}
	defer rows.Close()

	var templates []*repo.PageTemplate
	for rows.Next() {
		var (
			tpl       repo.PageTemplate
			projectNS sql.NullString
			created   int64
			updated   int64
		)
		if err := rows.Scan(&tpl.ID, &projectNS, &tpl.Name, &tpl.Description, &tpl.TemplateJSON, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan page template: %w", err)
		}
		tpl.ProjectID = scanNullableString(projectNS)
		tpl.CreatedAt = fromUnix(created)
		tpl.UpdatedAt = fromUnix(updated)
		templates = append(templates, &tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project page templates: %w", err)
	}
	return templates, nil
}

func (r *pageTemplateRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM page_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete page template: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
