package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

type imageSequenceRepo struct {
	db *sql.DB
}

func (r *imageSequenceRepo) Create(sequence *repo.ImageSequence) error {
	if sequence == nil {
		return fmt.Errorf("sequence is nil")
	}
	if sequence.ID == "" {
		sequence.ID = generateID("seq")
	}
	now := time.Now().UTC()
	if sequence.CreatedAt.IsZero() {
		sequence.CreatedAt = now
	}
	if sequence.UpdatedAt.IsZero() {
		sequence.UpdatedAt = sequence.CreatedAt
	}
	_, err := r.db.Exec(`INSERT INTO image_sequences (id, project_id, name, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		sequence.ID,
		sequence.ProjectID,
		sequence.Name,
		sequence.Description,
		toUnix(sequence.CreatedAt),
		toUnix(sequence.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert image sequence: %w", err)
	}
	return nil
}

func (r *imageSequenceRepo) Update(sequence *repo.ImageSequence) error {
	if sequence == nil {
		return fmt.Errorf("sequence is nil")
	}
	if sequence.UpdatedAt.IsZero() {
		sequence.UpdatedAt = time.Now().UTC()
	}
	res, err := r.db.Exec(`UPDATE image_sequences SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		sequence.Name,
		sequence.Description,
		toUnix(sequence.UpdatedAt),
		sequence.ID,
	)
	if err != nil {
		return fmt.Errorf("update image sequence: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *imageSequenceRepo) Get(id string) (*repo.ImageSequence, error) {
	row := r.db.QueryRow(`SELECT id, project_id, name, description, created_at, updated_at FROM image_sequences WHERE id = ?`, id)
	var (
		sequence repo.ImageSequence
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

func (r *imageSequenceRepo) ListByProject(projectID string) ([]*repo.ImageSequence, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, description, created_at, updated_at
        FROM image_sequences WHERE project_id = ?
        ORDER BY updated_at DESC, name ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list image sequences: %w", err)
	}
	defer rows.Close()

	var sequences []*repo.ImageSequence
	for rows.Next() {
		var (
			sequence repo.ImageSequence
			created  int64
			updated  int64
		)
		if err := rows.Scan(&sequence.ID, &sequence.ProjectID, &sequence.Name, &sequence.Description, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan image sequence: %w", err)
		}
		sequence.CreatedAt = fromUnix(created)
		sequence.UpdatedAt = fromUnix(updated)
		sequences = append(sequences, &sequence)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate image sequences: %w", err)
	}
	return sequences, nil
}

func (r *imageSequenceRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM image_sequences WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete image sequence: %w", err)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *imageSequenceRepo) AddItem(item *repo.ImageSequenceItem) error {
	if item == nil {
		return fmt.Errorf("sequence item is nil")
	}
	if !item.IsGap && (item.AssetID == nil || *item.AssetID == "") {
		return fmt.Errorf("asset id required for non-gap sequence item")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin add sequence item tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var nextPosition sql.NullInt64
	if err = tx.QueryRow(`SELECT MAX(position) FROM image_sequence_items WHERE sequence_id = ?`, item.SequenceID).Scan(&nextPosition); err != nil {
		return fmt.Errorf("query next position: %w", err)
	}
	position := 0
	if nextPosition.Valid {
		position = int(nextPosition.Int64) + 1
	}
	item.Position = position

	if _, err = tx.Exec(`INSERT INTO image_sequence_items (sequence_id, position, asset_id, is_gap)
        VALUES (?, ?, ?, ?)`,
		item.SequenceID,
		item.Position,
		nullStringPtr(item.AssetID),
		boolToInt(item.IsGap),
	); err != nil {
		return fmt.Errorf("insert image sequence item: %w", err)
	}

	return nil
}

func (r *imageSequenceRepo) ListItems(sequenceID string) ([]*repo.ImageSequenceItem, error) {
	rows, err := r.db.Query(`SELECT sequence_id, position, asset_id, is_gap
        FROM image_sequence_items WHERE sequence_id = ? ORDER BY position ASC`, sequenceID)
	if err != nil {
		return nil, fmt.Errorf("list image sequence items: %w", err)
	}
	defer rows.Close()

	var items []*repo.ImageSequenceItem
	for rows.Next() {
		var (
			item    repo.ImageSequenceItem
			assetID sql.NullString
			isGap   int
		)
		if err := rows.Scan(&item.SequenceID, &item.Position, &assetID, &isGap); err != nil {
			return nil, fmt.Errorf("scan image sequence item: %w", err)
		}
		item.AssetID = scanNullableString(assetID)
		item.IsGap = isGap == 1
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate image sequence items: %w", err)
	}
	return items, nil
}

func (r *imageSequenceRepo) ReplaceItems(sequenceID string, items []*repo.ImageSequenceItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin replace sequence items tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM image_sequence_items WHERE sequence_id = ?`, sequenceID); err != nil {
		return fmt.Errorf("clear sequence items: %w", err)
	}

	for idx, item := range items {
		if item == nil {
			return fmt.Errorf("sequence item at %d is nil", idx)
		}
		if !item.IsGap && (item.AssetID == nil || *item.AssetID == "") {
			return fmt.Errorf("asset id required for non-gap sequence item at position %d", idx)
		}
		if _, err = tx.Exec(`INSERT INTO image_sequence_items (sequence_id, position, asset_id, is_gap)
            VALUES (?, ?, ?, ?)`,
			sequenceID,
			idx,
			nullStringPtr(item.AssetID),
			boolToInt(item.IsGap),
		); err != nil {
			return fmt.Errorf("insert sequence item %d: %w", idx, err)
		}
	}

	return nil
}

func (r *imageSequenceRepo) DeleteItem(sequenceID string, position int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete sequence item tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	res, execErr := tx.Exec(`DELETE FROM image_sequence_items WHERE sequence_id = ? AND position = ?`, sequenceID, position)
	if execErr != nil {
		return fmt.Errorf("delete sequence item: %w", execErr)
	}
	if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows == 0 {
		return sql.ErrNoRows
	}

	rows, queryErr := tx.Query(`SELECT position FROM image_sequence_items WHERE sequence_id = ? ORDER BY position ASC`, sequenceID)
	if queryErr != nil {
		return fmt.Errorf("query sequence items for reindex: %w", queryErr)
	}
	defer rows.Close()

	newPos := 0
	for rows.Next() {
		var current int
		if scanErr := rows.Scan(&current); scanErr != nil {
			return fmt.Errorf("scan current position: %w", scanErr)
		}
		if current != newPos {
			if _, updateErr := tx.Exec(`UPDATE image_sequence_items SET position = ? WHERE sequence_id = ? AND position = ?`, newPos, sequenceID, current); updateErr != nil {
				return fmt.Errorf("reindex sequence items: %w", updateErr)
			}
		}
		newPos++
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("iterate positions: %w", err)
	}

	return nil
}
