package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"gitgud/internal/domain"
)

type RepoRepo struct {
	db *sql.DB
}

func NewRepoRepo(db *sql.DB) *RepoRepo {
	return &RepoRepo{db: db}
}

func (r *RepoRepo) Create(ctx context.Context, repo *domain.Repository) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO repositories(owner_id,name,description,is_private,default_branch) VALUES(?,?,?,?,?)`,
		repo.OwnerID, repo.Name, repo.Description, repo.IsPrivate, repo.DefaultBranch)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	repo.ID, _ = res.LastInsertId()
	return nil
}

func (r *RepoRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM repositories WHERE id = ?`, id)
	return err
}

func (r *RepoRepo) ByOwnerAndName(ctx context.Context, owner, name string) (*domain.Repository, error) {
	const q = `SELECT r.id, r.owner_id, u.username, r.name, r.description, r.is_private, r.default_branch, r.created_at
FROM repositories r
JOIN users u ON u.id = r.owner_id
WHERE u.username = ? AND r.name = ?`

	var repo domain.Repository
	err := r.db.QueryRowContext(ctx, q, owner, name).Scan(
		&repo.ID, &repo.OwnerID, &repo.OwnerName, &repo.Name, &repo.Description,
		&repo.IsPrivate, &repo.DefaultBranch, &repo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &repo, nil
}

func (r *RepoRepo) ListByOwner(ctx context.Context, ownerID int64) ([]*domain.Repository, error) {
	const q = `SELECT r.id, r.owner_id, u.username, r.name, r.description, r.is_private, r.default_branch, r.created_at
FROM repositories r
JOIN users u ON u.id = r.owner_id
WHERE r.owner_id = ?
ORDER BY r.created_at DESC, r.id DESC`

	rows, err := r.db.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*domain.Repository
	for rows.Next() {
		var repo domain.Repository
		if err := rows.Scan(
			&repo.ID, &repo.OwnerID, &repo.OwnerName, &repo.Name, &repo.Description,
			&repo.IsPrivate, &repo.DefaultBranch, &repo.CreatedAt); err != nil {
			return nil, err
		}
		repos = append(repos, &repo)
	}
	return repos, rows.Err()
}
