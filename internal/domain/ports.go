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
	Merge(ctx context.Context, owner, name, base, head, message, authorName, authorEmail string) error
}

type IssueRepository interface {
	Create(ctx context.Context, i *Issue) error
	ByNumber(ctx context.Context, repoID int64, number int) (*Issue, error)
	List(ctx context.Context, repoID int64, state IssueState) ([]*Issue, error)
	SetState(ctx context.Context, id int64, state IssueState) error

	AddComment(ctx context.Context, c *IssueComment) error
	Comments(ctx context.Context, issueID int64) ([]*IssueComment, error)
	CountByState(ctx context.Context, repoID int64) (open, closed int, err error)
}

type GitReader interface {
	IsEmpty(ctx context.Context, owner, name string) (bool, error)
	Branches(ctx context.Context, owner, name string) ([]string, error)
	Tip(ctx context.Context, owner, name, ref string) (*Commit, error)
	Tree(ctx context.Context, owner, name, ref, path string) ([]TreeEntry, error)
	Blob(ctx context.Context, owner, name, ref, path string) (*FileBlob, error)
	Log(ctx context.Context, owner, name, ref string, limit, offset int) ([]Commit, error)
	CommitDiff(ctx context.Context, owner, name, hash string) (*Commit, []FileDiff, error)
	Compare(ctx context.Context, owner, name, base, head string) (*Comparison, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *PullRequest) error
	ByNumber(ctx context.Context, repoID int64, number int) (*PullRequest, error)
	List(ctx context.Context, repoID int64, state PRState) ([]*PullRequest, error)
	SetState(ctx context.Context, id int64, state PRState) error

	AddComment(ctx context.Context, c *PRComment) error
	Comments(ctx context.Context, prID int64) ([]*PRComment, error)
	CountByState(ctx context.Context, repoID int64) (open, merged, closed int, err error)
}
