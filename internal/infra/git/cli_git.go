package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func (g CLIGit) RemovePath(ctx context.Context, owner, name string) error {
	path, err := g.repoPath(owner, name)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (g CLIGit) repoPath(owner, name string) (string, error) {
	if err := safeSegment(owner); err != nil {
		return "", err
	}
	if err := safeSegment(name); err != nil {
		return "", err
	}
	return filepath.Join(g.reposDir, owner, name+".git"), nil
}

func safeSegment(s string) error {
	if s == "" || s == "." || s == ".." ||
		strings.HasPrefix(s, ".") || strings.ContainsAny(s, `/\`) {
		return fmt.Errorf("invalid path segment %q", s)
	}
	return nil
}
