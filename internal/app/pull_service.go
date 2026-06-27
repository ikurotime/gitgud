package app

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gitgud/internal/domain"
)

type PullService struct {
	prs    domain.PullRequestRepository
	reader domain.GitReader
	git    domain.GitService
}

func NewPullService(prs domain.PullRequestRepository, reader domain.GitReader, git domain.GitService) *PullService {
	return &PullService{prs: prs, reader: reader, git: git}
}

func (s *PullService) Open(ctx context.Context, repo *domain.Repository, author *domain.User, title, body, base, head string) (*domain.PullRequest, error) {
	if author == nil {
		return nil, domain.ErrUnauthorized
	}
	title = strings.TrimSpace(title)
	base = strings.TrimSpace(base)
	head = strings.TrimSpace(head)

	if title == "" {
		return nil, fmt.Errorf("title is required: %w", domain.ErrValidation)
	}
	if base == "" || head == "" || base == head {
		return nil, fmt.Errorf("base and head must be two different branches: %w", domain.ErrValidation)
	}

	branches, err := s.reader.Branches(ctx, repo.OwnerName, repo.Name)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(branches, base) || !slices.Contains(branches, head) {
		return nil, fmt.Errorf("base and head must be existing branches: %w", domain.ErrValidation)
	}

	pr := &domain.PullRequest{
		RepoID:     repo.ID,
		AuthorID:   author.ID,
		Title:      title,
		Body:       strings.TrimSpace(body),
		BaseBranch: base,
		HeadBranch: head,
		State:      domain.PROpen,
	}
	if err := s.prs.Create(ctx, pr); err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *PullService) Compare(ctx context.Context, repo *domain.Repository, base, head string) (*domain.Comparison, error) {
	return s.reader.Compare(ctx, repo.OwnerName, repo.Name, base, head)
}

func (s *PullService) Comment(ctx context.Context, pr *domain.PullRequest, author *domain.User, body string) error {
	if author == nil {
		return domain.ErrUnauthorized
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("comment is required: %w", domain.ErrValidation)
	}
	return s.prs.AddComment(ctx, &domain.PRComment{
		PRID:     pr.ID,
		AuthorID: author.ID,
		Body:     body,
	})
}

func (s *PullService) Merge(ctx context.Context, repo *domain.Repository, pr *domain.PullRequest, actor *domain.User) error {
	if !CanMergePR(repo, actor) {
		return domain.ErrPermission
	}
	if pr.State != domain.PROpen {
		return fmt.Errorf("pull request is not open: %w", domain.ErrValidation)
	}

	msg := fmt.Sprintf("Merge pull request #%d from %s", pr.Number, pr.HeadBranch)
	if err := s.git.Merge(ctx, repo.OwnerName, repo.Name, pr.BaseBranch, pr.HeadBranch, msg, actor.Username, actor.Email); err != nil {
		return err
	}
	return s.prs.SetState(ctx, pr.ID, domain.PRMerged)
}

func (s *PullService) Close(ctx context.Context, repo *domain.Repository, pr *domain.PullRequest, actor *domain.User) error {
	if !(actor != nil && (actor.ID == pr.AuthorID || actor.ID == repo.OwnerID)) {
		return domain.ErrPermission
	}
	return s.prs.SetState(ctx, pr.ID, domain.PRClosed)
}

func (s *PullService) List(ctx context.Context, repoID int64, state domain.PRState) ([]*domain.PullRequest, error) {
	return s.prs.List(ctx, repoID, state)
}

func (s *PullService) Get(ctx context.Context, repoID int64, number int) (*domain.PullRequest, error) {
	return s.prs.ByNumber(ctx, repoID, number)
}

func (s *PullService) Comments(ctx context.Context, prID int64) ([]*domain.PRComment, error) {
	return s.prs.Comments(ctx, prID)
}

func (s *PullService) Counts(ctx context.Context, repoID int64) (open, merged, closed int, err error) {
	return s.prs.CountByState(ctx, repoID)
}

func CanMergePR(repo *domain.Repository, actor *domain.User) bool {
	return actor != nil && actor.ID == repo.OwnerID
}
