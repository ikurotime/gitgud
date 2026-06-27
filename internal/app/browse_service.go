package app

import (
	"context"

	"gitgud/internal/domain"
)

type BrowseService struct {
	repos domain.RepositoryRepository
	git   domain.GitReader
}

func NewBrowseService(repos domain.RepositoryRepository, git domain.GitReader) *BrowseService {
	return &BrowseService{repos: repos, git: git}
}

// Repo loads the repository and enforces the viewer's permission. The returned
// repo is safe to pass to the other methods, which assume it is authorized.
func (s *BrowseService) Repo(ctx context.Context, owner, name string, viewer *domain.User) (*domain.Repository, error) {
	repo, err := s.repos.ByOwnerAndName(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	if err := CanView(repo, viewer); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *BrowseService) ref(repo *domain.Repository, ref string) string {
	if ref == "" {
		return repo.DefaultBranch
	}
	return ref
}

func (s *BrowseService) IsEmpty(ctx context.Context, repo *domain.Repository) (bool, error) {
	return s.git.IsEmpty(ctx, repo.OwnerName, repo.Name)
}

func (s *BrowseService) Branches(ctx context.Context, repo *domain.Repository) ([]string, error) {
	return s.git.Branches(ctx, repo.OwnerName, repo.Name)
}

func (s *BrowseService) Tip(ctx context.Context, repo *domain.Repository, ref string) (*domain.Commit, error) {
	return s.git.Tip(ctx, repo.OwnerName, repo.Name, s.ref(repo, ref))
}

func (s *BrowseService) Tree(ctx context.Context, repo *domain.Repository, ref, path string) ([]domain.TreeEntry, error) {
	return s.git.Tree(ctx, repo.OwnerName, repo.Name, s.ref(repo, ref), path)
}

func (s *BrowseService) Blob(ctx context.Context, repo *domain.Repository, ref, path string) (*domain.FileBlob, error) {
	return s.git.Blob(ctx, repo.OwnerName, repo.Name, s.ref(repo, ref), path)
}

func (s *BrowseService) Log(ctx context.Context, repo *domain.Repository, ref string, limit, offset int) ([]domain.Commit, error) {
	return s.git.Log(ctx, repo.OwnerName, repo.Name, s.ref(repo, ref), limit, offset)
}

func (s *BrowseService) CommitDiff(ctx context.Context, repo *domain.Repository, hash string) (*domain.Commit, []domain.FileDiff, error) {
	return s.git.CommitDiff(ctx, repo.OwnerName, repo.Name, hash)
}
