package skill

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/model"
)

// CreateParams holds the data needed to insert a new skill.
type CreateParams struct {
	ID            string
	Name          string
	Description   string
	CategoryID    string
	Tags          json.RawMessage
	OwnerID       string
	OwnerName     string
	SpaceID       string
	Visibility    model.Visibility
	Version       string
	ReadmeContent string
	FileName      string
	FileURL       string
	FileSize      int64
	FileSHA256    string
}

// Create inserts a new skill record.
func (r *Repo) Create(ctx context.Context, p CreateParams) (*SkillRow, error) {
	now := time.Now().UTC()
	query := `
		INSERT INTO skills (id, name, description, category_id, tags, owner_id, owner_name,
			space_id, visibility, version, readme_content, file_name, file_url, file_size,
			file_sha256, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.CategoryID, string(p.Tags),
		p.OwnerID, p.OwnerName, p.SpaceID, string(p.Visibility), p.Version,
		p.ReadmeContent, p.FileName, p.FileURL, p.FileSize, p.FileSHA256,
		now, now,
	)
	if err != nil {
		return nil, err
	}
	return &SkillRow{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		CategoryID:    p.CategoryID,
		Tags:          p.Tags,
		OwnerID:       p.OwnerID,
		OwnerName:     p.OwnerName,
		SpaceID:       p.SpaceID,
		Visibility:    string(p.Visibility),
		Version:       p.Version,
		ReadmeContent: p.ReadmeContent,
		FileName:      p.FileName,
		FileURL:       p.FileURL,
		FileSize:      p.FileSize,
		FileSHA256:    p.FileSHA256,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}
