package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type layoutSequenceRepo struct {
	db *sql.DB
}

func (r *layoutSequenceRepo) Create(sequence *repo.LayoutSequence) error {
	if sequence == nil {
		return fmt.Errorf("layout sequence is nil")
	}
	if sequence.ID == "" {
		sequence.ID = generateID("lseq")
	}
	now := time.Now().UTC()
	if sequence.CreatedAt.IsZero() {
		sequence.CreatedAt = now
	}
	if sequence.UpdatedAt.IsZero() {
		sequence.UpdatedAt = sequence.CreatedAt
	}

	_, err := r.db.Exec(`INSERT INTO layout_sequences
        (id, project_id, name, description, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`,
		sequence.ID,
		sequence.ProjectID,
		sequence.Name,
		sequence.Description,
		toUnix(sequence.CreatedAt),
		toUnix(sequence.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert layout sequence: %w", err)
	}
	return nil
}

func (r *layoutSequenceRepo) Update(sequence *repo.LayoutSequence) error {
	if sequence == nil {
		return fmt.Errorf("layout sequence is nil")
	}
	if sequence.UpdatedAt.IsZero() {
		sequence.UpdatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(`UPDATE layout_sequences
        SET name = ?, description = ?, updated_at = ?
        WHERE id = ?`,
		sequence.Name,
		sequence.Description,
		toUnix(sequence.UpdatedAt),
		sequence.ID,
	)
	if err != nil {
		return fmt.Errorf("update layout sequence: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *layoutSequenceRepo) Get(id string) (*repo.LayoutSequence, error) {
	row := r.db.QueryRow(`SELECT id, project_id, name, description, created_at, updated_at
        FROM layout_sequences WHERE id = ?`, id)
	var (
		sequence repo.LayoutSequence
		created  int64
		updated  int64
	)
	if err := row.Scan(&sequence.ID, &sequence.ProjectID, &sequence.Name, &sequence.Description, &created, &updated); err != nil {
		return nil, errNotFound(err)
	}
	sequence.CreatedAt = fromUnix(created)
	sequence.UpdatedAt = fromUnix(updated)
	return &sequence, nil
}

func (r *layoutSequenceRepo) ListByProject(projectID string) ([]*repo.LayoutSequence, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, created_at, updated_at
        FROM layout_sequences WHERE project_id = ?
        ORDER BY updated_at DESC, name ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list layout sequences: %w", err)
	}
	defer rows.Close()

	var sequences []*repo.LayoutSequence
	for rows.Next() {
		var (
			sequence repo.LayoutSequence
			created  int64
			updated  int64
		)
		if err := rows.Scan(&sequence.ID, &sequence.ProjectID, &sequence.Name, &sequence.Description, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan layout sequence: %w", err)
		}
		sequence.CreatedAt = fromUnix(created)
		sequence.UpdatedAt = fromUnix(updated)
		sequences = append(sequences, &sequence)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout sequences: %w", err)
	}
	return sequences, nil
}

func (r *layoutSequenceRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM layout_sequences WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete layout sequence: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *layoutSequenceRepo) AddItem(item *repo.LayoutSequenceItem) error {
	if item == nil {
		return fmt.Errorf("layout sequence item is nil")
	}
	if item.LaidOutImageID == "" {
		return fmt.Errorf("laid-out image id required")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin add layout sequence item tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var nextPosition sql.NullInt64
	if err = tx.QueryRow(`SELECT MAX(position) FROM layout_sequence_items WHERE sequence_id = ?`, item.SequenceID).Scan(&nextPosition); err != nil {
		return fmt.Errorf("query next layout sequence position: %w", err)
	}
	position := 0
	if nextPosition.Valid {
		position = int(nextPosition.Int64) + 1
	}
	item.Position = position

	if _, err = tx.Exec(`INSERT INTO layout_sequence_items (sequence_id, position, laid_out_image_id)
        VALUES (?, ?, ?)`,
		item.SequenceID,
		item.Position,
		item.LaidOutImageID,
	); err != nil {
		return fmt.Errorf("insert layout sequence item: %w", err)
	}
	return nil
}

func (r *layoutSequenceRepo) ListItems(sequenceID string) ([]*repo.LayoutSequenceItem, error) {
	rows, err := r.db.Query(`SELECT sequence_id, position, laid_out_image_id
        FROM layout_sequence_items WHERE sequence_id = ?
        ORDER BY position ASC`, sequenceID)
	if err != nil {
		return nil, fmt.Errorf("list layout sequence items: %w", err)
	}
	defer rows.Close()

	var items []*repo.LayoutSequenceItem
	for rows.Next() {
		var item repo.LayoutSequenceItem
		if err := rows.Scan(&item.SequenceID, &item.Position, &item.LaidOutImageID); err != nil {
			return nil, fmt.Errorf("scan layout sequence item: %w", err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout sequence items: %w", err)
	}
	return items, nil
}

func (r *layoutSequenceRepo) ReplaceItems(sequenceID string, items []*repo.LayoutSequenceItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin replace layout sequence items tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM layout_sequence_items WHERE sequence_id = ?`, sequenceID); err != nil {
		return fmt.Errorf("delete existing layout sequence items: %w", err)
	}

	for _, item := range items {
		if item.LaidOutImageID == "" {
			return fmt.Errorf("laid-out image id required")
		}
		if _, err = tx.Exec(`INSERT INTO layout_sequence_items (sequence_id, position, laid_out_image_id)
            VALUES (?, ?, ?)`,
			sequenceID,
			item.Position,
			item.LaidOutImageID,
		); err != nil {
			return fmt.Errorf("insert layout sequence item: %w", err)
		}
	}

	return nil
}

func (r *layoutSequenceRepo) DeleteItem(sequenceID string, position int) error {
	res, err := r.db.Exec(`DELETE FROM layout_sequence_items WHERE sequence_id = ? AND position = ?`, sequenceID, position)
	if err != nil {
		return fmt.Errorf("delete layout sequence item: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}
