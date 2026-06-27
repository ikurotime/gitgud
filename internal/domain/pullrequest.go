package domain

import "time"

type PRState string

const (
	PROpen   PRState = "open"
	PRMerged PRState = "merged"
	PRClosed PRState = "closed"
)

type PullRequest struct {
	ID         int64
	RepoID     int64
	Number     int
	AuthorID   int64
	AuthorName string
	Title      string
	Body       string
	BaseBranch string
	HeadBranch string
	State      PRState
	CreatedAt  time.Time
}

type PRComment struct {
	ID         int64
	PRID       int64
	AuthorID   int64
	AuthorName string
	Body       string
	CreatedAt  time.Time
}

type Comparison struct {
	Commits   []Commit
	Files     []FileDiff
	Mergeable bool
}
