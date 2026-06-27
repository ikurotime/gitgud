package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"gitgud/internal/domain"
)

type PRRepo struct {
	db *sql.DB
}

func NewPRRepo(db *sql.DB) *PRRepo {
	return &PRRepo{db: db}
}

func (r *PRRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var next int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(number),0)+1 FROM pull_requests WHERE repo_id=?`, pr.RepoID).Scan(&next)
	if err != nil {
		return err
	}
	pr.Number = next

	res, err := tx.ExecContext(ctx,
		`INSERT INTO pull_requests(repo_id,number,author_id,title,body,base_branch,head_branch,state)
		 VALUES(?,?,?,?,?,?,?,?)`,
		pr.RepoID, pr.Number, pr.AuthorID, pr.Title, pr.Body, pr.BaseBranch, pr.HeadBranch, string(pr.State))
	if err != nil {
		return err
	}
	pr.ID, _ = res.LastInsertId()
	return tx.Commit()
}

func (r *PRRepo) ByNumber(ctx context.Context, repoID int64, number int) (*domain.PullRequest, error) {
	const q = `SELECT p.id, p.repo_id, p.number, p.author_id, u.username, p.title, p.body, p.base_branch, p.head_branch, p.state, p.created_at
FROM pull_requests p
JOIN users u ON u.id = p.author_id
WHERE p.repo_id = ? AND p.number = ?`

	pr, err := scanPR(r.db.QueryRowContext(ctx, q, repoID, number))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return pr, nil
}

func (r *PRRepo) List(ctx context.Context, repoID int64, state domain.PRState) ([]*domain.PullRequest, error) {
	q := `SELECT p.id, p.repo_id, p.number, p.author_id, u.username, p.title, p.body, p.base_branch, p.head_branch, p.state, p.created_at
FROM pull_requests p
JOIN users u ON u.id = p.author_id
WHERE p.repo_id = ?`
	args := []any{repoID}
	if state != "" {
		q += ` AND p.state = ?`
		args = append(args, string(state))
	}
	q += ` ORDER BY p.number DESC`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		pr, err := scanPR(rows)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (r *PRRepo) SetState(ctx context.Context, id int64, state domain.PRState) error {
	_, err := r.db.ExecContext(ctx, `UPDATE pull_requests SET state=? WHERE id=?`, string(state), id)
	return err
}

func (r *PRRepo) AddComment(ctx context.Context, c *domain.PRComment) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO pr_comments(pr_id,author_id,body) VALUES(?,?,?)`,
		c.PRID, c.AuthorID, c.Body)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (r *PRRepo) Comments(ctx context.Context, prID int64) ([]*domain.PRComment, error) {
	const q = `SELECT c.id, c.pr_id, c.author_id, u.username, c.body, c.created_at
FROM pr_comments c
JOIN users u ON u.id = c.author_id
WHERE c.pr_id = ?
ORDER BY c.created_at, c.id`

	rows, err := r.db.QueryContext(ctx, q, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.PRComment
	for rows.Next() {
		var c domain.PRComment
		if err := rows.Scan(&c.ID, &c.PRID, &c.AuthorID, &c.AuthorName, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}

func (r *PRRepo) CountByState(ctx context.Context, repoID int64) (open, merged, closed int, err error) {
	const q = `SELECT
  COALESCE(SUM(CASE WHEN state='open' THEN 1 ELSE 0 END),0),
  COALESCE(SUM(CASE WHEN state='merged' THEN 1 ELSE 0 END),0),
  COALESCE(SUM(CASE WHEN state='closed' THEN 1 ELSE 0 END),0)
FROM pull_requests WHERE repo_id=?`
	err = r.db.QueryRowContext(ctx, q, repoID).Scan(&open, &merged, &closed)
	return open, merged, closed, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPR(row rowScanner) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.RepoID, &pr.Number, &pr.AuthorID, &pr.AuthorName,
		&pr.Title, &pr.Body, &pr.BaseBranch, &pr.HeadBranch, &pr.State, &pr.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}
