package app

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gitgud/internal/domain"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9-]{1,39}$`)

type UserService struct {
	users  domain.UserRepository
	hasher domain.PasswordHasher
}

func NewUserService(u domain.UserRepository, h domain.PasswordHasher) *UserService {
	return &UserService{users: u, hasher: h}
}

func (s *UserService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.TrimSpace(email)

	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email and password are required: %w", domain.ErrValidation)
	}
	if !usernameRe.MatchString(username) {
		return nil, fmt.Errorf("username must be 1-39 chars of a-z, 0-9 or -: %w", domain.ErrValidation)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters: %w", domain.ErrValidation)
	}

	if _, err := s.users.ByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("username %q is taken: %w", username, domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	u, err := s.users.ByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return nil, domain.ErrUnauthorized
	}
	return u, nil
}

func (s *UserService) ByID(ctx context.Context, id string) (*domain.User, error) {
	return s.users.ByID(ctx, id)
}

func (s *UserService) ByUsername(ctx context.Context, name string) (*domain.User, error) {
	return s.users.ByUsername(ctx, strings.ToLower(strings.TrimSpace(name)))
}
