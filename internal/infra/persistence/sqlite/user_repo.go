package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"

	"gitgud/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO users(username,email,password_hash) VALUES(?,?,?)`,
		u.Username, u.Email, u.PasswordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

func (r *UserRepo) ByUsername(ctx context.Context, name string) (*domain.User, error) {
	return r.queryUser(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE username = ?`,
		name)
}

func (r *UserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	return r.queryUser(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE id = ?`,
		id)
}

func (r *UserRepo) queryUser(ctx context.Context, query string, arg any) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) &&
		sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}
