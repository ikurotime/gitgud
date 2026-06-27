package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"gitgud/internal/domain"
)

type IssueRepo struct {
	db *sql.DB
}

func NewIssueRepo(db *sql.DB) *IssueRepo {
	return &IssueRepo{db: db}
}

func (r *IssueRepo) Create(ctx context.Context, i *domain.Issue) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var next int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(number),0)+1 FROM issues WHERE repo_id=?`, i.RepoID).Scan(&next)
	if err != nil {
		return err
	}
	i.Number = next

	res, err := tx.ExecContext(ctx,
		`INSERT INTO issues(repo_id,number,author_id,title,body,state) VALUES(?,?,?,?,?,?)`,
		i.RepoID, i.Number, i.AuthorID, i.Title, i.Body, string(i.State))
	if err != nil {
		return err
	}
	i.ID, _ = res.LastInsertId()
	return tx.Commit()
}

func (r *IssueRepo) ByNumber(ctx context.Context, repoID int64, number int) (*domain.Issue, error) {
	const q = `SELECT i.id, i.repo_id, i.number, i.author_id, u.username, i.title, i.body, i.state, i.created_at
FROM issues i
JOIN users u ON u.id = i.author_id
WHERE i.repo_id = ? AND i.number = ?`

	var i domain.Issue
	err := r.db.QueryRowContext(ctx, q, repoID, number).Scan(
		&i.ID, &i.RepoID, &i.Number, &i.AuthorID, &i.AuthorName, &i.Title, &i.Body, &i.State, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &i, nil
}

func (r *IssueRepo) List(ctx context.Context, repoID int64, state domain.IssueState) ([]*domain.Issue, error) {
	q := `SELECT i.id, i.repo_id, i.number, i.author_id, u.username, i.title, i.body, i.state, i.created_at
FROM issues i
JOIN users u ON u.id = i.author_id
WHERE i.repo_id = ?`
	args := []any{repoID}
	if state != "" {
		q += ` AND i.state = ?`
		args = append(args, string(state))
	}
	q += ` ORDER BY i.number DESC`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*domain.Issue
	for rows.Next() {
		var i domain.Issue
		if err := rows.Scan(&i.ID, &i.RepoID, &i.Number, &i.AuthorID, &i.AuthorName,
			&i.Title, &i.Body, &i.State, &i.CreatedAt); err != nil {
			return nil, err
		}
		issues = append(issues, &i)
	}
	return issues, rows.Err()
}

func (r *IssueRepo) SetState(ctx context.Context, id int64, state domain.IssueState) error {
	_, err := r.db.ExecContext(ctx, `UPDATE issues SET state=? WHERE id=?`, string(state), id)
	return err
}

func (r *IssueRepo) AddComment(ctx context.Context, c *domain.IssueComment) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO issue_comments(issue_id,author_id,body) VALUES(?,?,?)`,
		c.IssueID, c.AuthorID, c.Body)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (r *IssueRepo) Comments(ctx context.Context, issueID int64) ([]*domain.IssueComment, error) {
	const q = `SELECT c.id, c.issue_id, c.author_id, u.username, c.body, c.created_at
FROM issue_comments c
JOIN users u ON u.id = c.author_id
WHERE c.issue_id = ?
ORDER BY c.created_at, c.id`

	rows, err := r.db.QueryContext(ctx, q, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.IssueComment
	for rows.Next() {
		var c domain.IssueComment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.AuthorID, &c.AuthorName, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}

func (r *IssueRepo) CountByState(ctx context.Context, repoID int64) (open, closed int, err error) {
	const q = `SELECT
  COALESCE(SUM(CASE WHEN state='open' THEN 1 ELSE 0 END),0),
  COALESCE(SUM(CASE WHEN state='closed' THEN 1 ELSE 0 END),0)
FROM issues WHERE repo_id=?`
	err = r.db.QueryRowContext(ctx, q, repoID).Scan(&open, &closed)
	return open, closed, err
}
