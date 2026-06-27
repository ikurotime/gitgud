package app

import (
	"context"
	"fmt"
	"strings"

	"gitgud/internal/domain"
)

type IssueService struct {
	issues domain.IssueRepository
}

func NewIssueService(issues domain.IssueRepository) *IssueService {
	return &IssueService{issues: issues}
}

func (s *IssueService) Open(ctx context.Context, repo *domain.Repository, author *domain.User, title, body string) (*domain.Issue, error) {
	if author == nil {
		return nil, domain.ErrUnauthorized
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title is required: %w", domain.ErrValidation)
	}

	issue := &domain.Issue{
		RepoID:   repo.ID,
		AuthorID: author.ID,
		Title:    title,
		Body:     strings.TrimSpace(body),
		State:    domain.IssueOpen,
	}
	if err := s.issues.Create(ctx, issue); err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *IssueService) Comment(ctx context.Context, issue *domain.Issue, author *domain.User, body string) error {
	if author == nil {
		return domain.ErrUnauthorized
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("comment is required: %w", domain.ErrValidation)
	}
	return s.issues.AddComment(ctx, &domain.IssueComment{
		IssueID:  issue.ID,
		AuthorID: author.ID,
		Body:     body,
	})
}

func (s *IssueService) Close(ctx context.Context, repo *domain.Repository, issue *domain.Issue, actor *domain.User) error {
	if !canModifyIssue(repo, issue, actor) {
		return domain.ErrPermission
	}
	return s.issues.SetState(ctx, issue.ID, domain.IssueClosed)
}

func (s *IssueService) Reopen(ctx context.Context, repo *domain.Repository, issue *domain.Issue, actor *domain.User) error {
	if !canModifyIssue(repo, issue, actor) {
		return domain.ErrPermission
	}
	return s.issues.SetState(ctx, issue.ID, domain.IssueOpen)
}

func (s *IssueService) List(ctx context.Context, repoID int64, state domain.IssueState) ([]*domain.Issue, error) {
	return s.issues.List(ctx, repoID, state)
}

func (s *IssueService) Get(ctx context.Context, repoID int64, number int) (*domain.Issue, []*domain.IssueComment, error) {
	issue, err := s.issues.ByNumber(ctx, repoID, number)
	if err != nil {
		return nil, nil, err
	}
	comments, err := s.issues.Comments(ctx, issue.ID)
	if err != nil {
		return nil, nil, err
	}
	return issue, comments, nil
}

func (s *IssueService) Comments(ctx context.Context, issueID int64) ([]*domain.IssueComment, error) {
	return s.issues.Comments(ctx, issueID)
}

func (s *IssueService) Counts(ctx context.Context, repoID int64) (open, closed int, err error) {
	return s.issues.CountByState(ctx, repoID)
}

func CanModifyIssue(repo *domain.Repository, issue *domain.Issue, actor *domain.User) bool {
	return canModifyIssue(repo, issue, actor)
}

func canModifyIssue(repo *domain.Repository, issue *domain.Issue, actor *domain.User) bool {
	return actor != nil && (actor.ID == issue.AuthorID || actor.ID == repo.OwnerID)
}
