package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitgud/internal/domain"
)

type CLIGit struct {
	reposDir string
}

func NewCLIGit(reposDir string) CLIGit {
	return CLIGit{reposDir: reposDir}
}

func (g CLIGit) InitBare(ctx context.Context, owner, name, defaultBranch string) error {
	path, err := g.repoPath(owner, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "init", "--bare",
		"--initial-branch="+defaultBranch, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %v: %s", err, out)
	}

	cmd = exec.CommandContext(ctx, "git", "-C", path, "config", "http.receivepack", "true")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config http.receivepack: %v: %s", err, out)
	}
	return nil
}

func (g CLIGit) Merge(ctx context.Context, owner, name, base, head, message, authorName, authorEmail string) error {
	bare, err := g.repoPath(owner, name)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "gitgud-merge-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	run := func(args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = tmp
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME="+authorName, "GIT_AUTHOR_EMAIL="+authorEmail,
			"GIT_COMMITTER_NAME="+authorName, "GIT_COMMITTER_EMAIL="+authorEmail)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	if out, err := run("clone", bare, "."); err != nil {
		return fmt.Errorf("clone: %v: %s", err, out)
	}
	if out, err := run("checkout", base); err != nil {
		return fmt.Errorf("checkout %s: %v: %s", base, err, out)
	}
	if _, err := run("merge", "--no-ff", "-m", message, "origin/"+head); err != nil {
		return domain.ErrConflict
	}
	if out, err := run("push", "origin", base); err != nil {
		return fmt.Errorf("push: %v: %s", err, out)
	}
	return nil
}

func (g CLIGit) RemovePath(ctx context.Context, owner, name string) error {
	path, err := g.repoPath(owner, name)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (g CLIGit) repoPath(owner, name string) (string, error) {
	return repoPath(g.reposDir, owner, name)
}

func repoPath(reposDir, owner, name string) (string, error) {
	if err := safeSegment(owner); err != nil {
		return "", err
	}
	if err := safeSegment(name); err != nil {
		return "", err
	}
	return filepath.Join(reposDir, owner, name+".git"), nil
}

func safeSegment(s string) error {
	if s == "" || s == "." || s == ".." ||
		strings.HasPrefix(s, ".") || strings.ContainsAny(s, `/\`) {
		return fmt.Errorf("invalid path segment %q", s)
	}
	return nil
}
