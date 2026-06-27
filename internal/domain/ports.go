package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	ByUsername(ctx context.Context, name string) (*User, error)
	ByID(ctx context.Context, id string) (*User, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

type RepositoryRepository interface {
	Create(ctx context.Context, r *Repository) error
	Delete(ctx context.Context, id int64) error
	ByOwnerAndName(ctx context.Context, owner, name string) (*Repository, error)
	ListByOwner(ctx context.Context, ownerID int64) ([]*Repository, error)
}

type GitService interface {
	InitBare(ctx context.Context, owner, name, defaultBranch string) error
	RemovePath(ctx context.Context, owner, name string) error
}
