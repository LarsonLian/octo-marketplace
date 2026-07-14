package skill

import "database/sql"

// Repo provides data access for skills.
type Repo struct {
	db *sql.DB
}

// New creates a new skill repository.
func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}
