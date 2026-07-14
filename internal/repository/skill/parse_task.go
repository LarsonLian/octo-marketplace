package skill

import (
	"context"
	"database/sql"
	"encoding/json"
)

// ParseTaskRow holds parse_task data needed for skill creation.
type ParseTaskRow struct {
	ID                string
	UploadID          string
	FileURL           string
	Status            string
	ResultName        string
	ResultDescription *string
	ResultVersion     string
	ResultTags        json.RawMessage
	ResultReadme      *string
	OwnerID           string
	SpaceID           string
}

// GetParseTask retrieves a parse task by ID.
func (r *Repo) GetParseTask(ctx context.Context, id string) (*ParseTaskRow, error) {
	query := `
		SELECT id, upload_id, file_url, status, result_name, result_description,
			result_version, result_tags, result_readme, owner_id, space_id
		FROM parse_tasks
		WHERE id = ?
	`
	var pt ParseTaskRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pt.ID, &pt.UploadID, &pt.FileURL, &pt.Status,
		&pt.ResultName, &pt.ResultDescription, &pt.ResultVersion,
		&pt.ResultTags, &pt.ResultReadme, &pt.OwnerID, &pt.SpaceID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

// MarkParseTaskConsumed marks a parse task as consumed (changes status to prevent reuse).
func (r *Repo) MarkParseTaskConsumed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE parse_tasks SET status = 'consumed' WHERE id = ?", id)
	return err
}
