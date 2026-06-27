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
