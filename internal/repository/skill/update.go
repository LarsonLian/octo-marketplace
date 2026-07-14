package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
)

// UpdateParams holds optional fields to update.
type UpdateParams struct {
	Name        *string
	Description *string
	CategoryID  *string
	Tags        json.RawMessage // nil means no change
	Visibility  *model.Visibility
	Version     *string
}

// Update updates the specified fields on a skill. Returns the number of affected rows.
func (r *Repo) Update(ctx context.Context, id string, p UpdateParams) (int64, error) {
	var sets []string
	var args []interface{}

	if p.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *p.Name)
	}
	if p.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *p.Description)
	}
	if p.CategoryID != nil {
		sets = append(sets, "category_id = ?")
		args = append(args, *p.CategoryID)
	}
	if p.Tags != nil {
		sets = append(sets, "tags = ?")
		args = append(args, string(p.Tags))
	}
	if p.Visibility != nil {
		sets = append(sets, "visibility = ?")
		args = append(args, string(*p.Visibility))
	}
	if p.Version != nil {
		sets = append(sets, "version = ?")
		args = append(args, *p.Version)
	}

	if len(sets) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf("UPDATE skills SET %s WHERE id = ?", strings.Join(sets, ", "))
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
