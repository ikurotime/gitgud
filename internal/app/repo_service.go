package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gitgud/internal/domain"
)

var repoNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-_.]{0,99}$`)

type RepoService struct {
	repos domain.RepositoryRepository
	git   domain.GitService
}

func NewRepoService(repos domain.RepositoryRepository, git domain.GitService) *RepoService {
	return &RepoService{repos: repos, git: git}
}

func (s *RepoService) CreateRepo(ctx context.Context, owner *domain.User, name, description string, private bool) (*domain.Repository, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "." || name == ".." || !repoNameRe.MatchString(name) {
		return nil, fmt.Errorf("name must be 1-100 chars of a-z, 0-9, -, _ or . and start with a letter or number: %w", domain.ErrValidation)
	}

	repo := &domain.Repository{
		OwnerID:       owner.ID,
		OwnerName:     owner.Username,
		Name:          name,
		Description:   strings.TrimSpace(description),
		IsPrivate:     private,
		DefaultBranch: "main",
	}

	if err := s.repos.Create(ctx, repo); err != nil {
		return nil, err
	}

	if err := s.git.InitBare(ctx, owner.Username, name, repo.DefaultBranch); err != nil {
		_ = s.repos.Delete(ctx, repo.ID)
		return nil, fmt.Errorf("init repository on disk: %w", err)
	}

	return repo, nil
}

func (s *RepoService) Get(ctx context.Context, owner, name string) (*domain.Repository, error) {
	return s.repos.ByOwnerAndName(ctx, owner, name)
}

func (s *RepoService) ListByOwner(ctx context.Context, ownerID int64) ([]*domain.Repository, error) {
	return s.repos.ListByOwner(ctx, ownerID)
}

func CanView(repo *domain.Repository, viewer *domain.User) error {
	if !repo.IsPrivate {
		return nil
	}
	if viewer != nil && viewer.ID == repo.OwnerID {
		return nil
	}
	return domain.ErrNotFound
}

func CanPush(repo *domain.Repository, viewer *domain.User) bool {
	return viewer != nil && viewer.ID == repo.OwnerID
}
