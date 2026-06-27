package domain

import "time"

type Repository struct {
	ID            int64
	OwnerID       int64
	OwnerName     string
	Name          string
	Description   string
	IsPrivate     bool
	DefaultBranch string
	CreatedAt     time.Time
}
