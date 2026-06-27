package app

import (
	"context"

	"gitgud/internal/domain"
)

type GitAccessService struct {
	repos domain.RepositoryRepository
}

func NewGitAccessService(repos domain.RepositoryRepository) *GitAccessService {
	return &GitAccessService{repos: repos}
}

func (s *GitAccessService) Authorize(ctx context.Context, owner, name string, push bool, user *domain.User) (allow, needAuth bool, remoteUser string, err error) {
	repo, err := s.repos.ByOwnerAndName(ctx, owner, name)
	if err != nil {
		return false, false, "", err
	}

	if push {
		if user == nil {
			return false, true, "", nil
		}
		if !CanPush(repo, user) {
			return false, false, "", domain.ErrPermission
		}
	} else if repo.IsPrivate {
		if user == nil {
			return false, true, "", nil
		}
		if err := CanView(repo, user); err != nil {
			return false, false, "", err
		}
	}

	if user != nil {
		remoteUser = user.Username
	}
	return true, false, remoteUser, nil
}
