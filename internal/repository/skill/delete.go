package skill

import "context"

// Delete hard-deletes a skill by ID. Returns the number of affected rows.
func (r *Repo) Delete(ctx context.Context, id string) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM skills WHERE id = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
