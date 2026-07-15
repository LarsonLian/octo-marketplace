package model

import (
	"encoding/json"
	"time"
)

// Visibility defines the access scope of a skill.
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilitySpace   Visibility = "space"
	VisibilityPrivate Visibility = "private"
)

// Skill represents a published marketplace skill.
type Skill struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	CategoryID    string          `json:"category_id"`
	Tags          json.RawMessage `json:"tags"`
	OwnerID       string          `json:"owner_id"`
	OwnerName     string          `json:"owner_name"`
	SpaceID       string          `json:"space_id"`
	Visibility    Visibility      `json:"visibility"`
	Version       string          `json:"version"`
	ReadmeContent string          `json:"readme_content"`
	FileName      string          `json:"file_name"`
	FileURL       string          `json:"file_url"`
	FileSize      int64           `json:"file_size"`
	FileSHA256    string          `json:"file_sha256"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
