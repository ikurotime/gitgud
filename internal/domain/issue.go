package domain

import "time"

type IssueState string

const (
	IssueOpen   IssueState = "open"
	IssueClosed IssueState = "closed"
)

type Issue struct {
	ID         int64
	RepoID     int64
	Number     int
	AuthorID   int64
	AuthorName string
	Title      string
	Body       string
	State      IssueState
	CreatedAt  time.Time
}

type IssueComment struct {
	ID         int64
	IssueID    int64
	AuthorID   int64
	AuthorName string
	Body       string
	CreatedAt  time.Time
}
