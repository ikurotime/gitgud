package git

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"gitgud/internal/domain"
)

type GoGitReader struct {
	reposDir string
}

func NewGoGitReader(reposDir string) *GoGitReader {
	return &GoGitReader{reposDir: reposDir}
}

func (g *GoGitReader) open(owner, name string) (*gogit.Repository, error) {
	p, err := repoPath(g.reposDir, owner, name)
	if err != nil {
		return nil, err
	}
	repo, err := gogit.PlainOpen(p)
	if err != nil {
		return nil, mapErr(err)
	}
	return repo, nil
}

func (g *GoGitReader) resolveCommit(repo *gogit.Repository, ref string) (*object.Commit, error) {
	if strings.TrimSpace(ref) == "" {
		ref = "HEAD"
	}
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, mapErr(err)
	}
	c, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

func (g *GoGitReader) IsEmpty(ctx context.Context, owner, name string) (bool, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return false, err
	}
	_, err = repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return true, nil
	}
	if err != nil {
		return false, mapErr(err)
	}
	return false, nil
}

func (g *GoGitReader) Branches(ctx context.Context, owner, name string) ([]string, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, err
	}
	iter, err := repo.Branches()
	if err != nil {
		return nil, mapErr(err)
	}
	var names []string
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		names = append(names, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, mapErr(err)
	}
	sort.Strings(names)
	return names, nil
}

func (g *GoGitReader) Tip(ctx context.Context, owner, name, ref string) (*domain.Commit, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, err
	}
	c, err := g.resolveCommit(repo, ref)
	if err != nil {
		return nil, err
	}
	return commitDTO(c), nil
}

func (g *GoGitReader) Tree(ctx context.Context, owner, name, ref, p string) ([]domain.TreeEntry, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, err
	}
	c, err := g.resolveCommit(repo, ref)
	if err != nil {
		return nil, err
	}
	tree, err := c.Tree()
	if err != nil {
		return nil, mapErr(err)
	}
	p = strings.Trim(p, "/")
	if p != "" {
		tree, err = tree.Tree(p)
		if err != nil {
			return nil, mapErr(err)
		}
	}

	entries := make([]domain.TreeEntry, 0, len(tree.Entries))
	for _, e := range tree.Entries {
		isDir := e.Mode == filemode.Dir
		var size int64
		if !isDir {
			if f, err := tree.TreeEntryFile(&e); err == nil {
				size = f.Size
			}
		}
		entries = append(entries, domain.TreeEntry{
			Name:  e.Name,
			Path:  path.Join(p, e.Name),
			IsDir: isDir,
			Size:  size,
			Mode:  e.Mode.String(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func (g *GoGitReader) Blob(ctx context.Context, owner, name, ref, p string) (*domain.FileBlob, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, err
	}
	c, err := g.resolveCommit(repo, ref)
	if err != nil {
		return nil, err
	}
	tree, err := c.Tree()
	if err != nil {
		return nil, mapErr(err)
	}
	f, err := tree.File(strings.Trim(p, "/"))
	if err != nil {
		return nil, mapErr(err)
	}
	reader, err := f.Reader()
	if err != nil {
		return nil, mapErr(err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, mapErr(err)
	}
	return &domain.FileBlob{
		Path:     strings.Trim(p, "/"),
		Content:  content,
		IsBinary: isBinary(content),
		Size:     f.Size,
	}, nil
}

func (g *GoGitReader) Log(ctx context.Context, owner, name, ref string, limit, offset int) ([]domain.Commit, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, err
	}
	c, err := g.resolveCommit(repo, ref)
	if err != nil {
		return nil, err
	}
	iter, err := repo.Log(&gogit.LogOptions{From: c.Hash})
	if err != nil {
		return nil, mapErr(err)
	}
	defer iter.Close()

	var commits []domain.Commit
	skipped := 0
	err = iter.ForEach(func(commit *object.Commit) error {
		if skipped < offset {
			skipped++
			return nil
		}
		if limit > 0 && len(commits) >= limit {
			return storerStop
		}
		commits = append(commits, *commitDTO(commit))
		return nil
	})
	if err != nil && !errors.Is(err, storerStop) {
		return nil, mapErr(err)
	}
	return commits, nil
}

func (g *GoGitReader) CommitDiff(ctx context.Context, owner, name, hash string) (*domain.Commit, []domain.FileDiff, error) {
	repo, err := g.open(owner, name)
	if err != nil {
		return nil, nil, err
	}
	c, err := repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, nil, mapErr(err)
	}

	thisTree, err := c.Tree()
	if err != nil {
		return nil, nil, mapErr(err)
	}

	var parentTree *object.Tree
	if parent, err := c.Parents().Next(); err == nil {
		parentTree, err = parent.Tree()
		if err != nil {
			return nil, nil, mapErr(err)
		}
	} else if !errors.Is(err, io.EOF) {
		return nil, nil, mapErr(err)
	}

	changes, err := object.DiffTree(parentTree, thisTree)
	if err != nil {
		return nil, nil, mapErr(err)
	}
	patch, err := changes.Patch()
	if err != nil {
		return nil, nil, mapErr(err)
	}

	var diffs []domain.FileDiff
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		p := ""
		if to != nil {
			p = to.Path()
		} else if from != nil {
			p = from.Path()
		}

		var sb strings.Builder
		added, deleted := 0, 0
		for _, chunk := range fp.Chunks() {
			lines := strings.SplitAfter(chunk.Content(), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				switch chunk.Type() {
				case 1: // Add
					sb.WriteString("+" + line)
					added++
				case 2: // Delete
					sb.WriteString("-" + line)
					deleted++
				default: // Equal
					sb.WriteString(" " + line)
				}
			}
		}
		diffs = append(diffs, domain.FileDiff{
			Path:    p,
			Patch:   sb.String(),
			Added:   added,
			Deleted: deleted,
		})
	}

	return commitDTO(c), diffs, nil
}

var storerStop = errors.New("stop")

func commitDTO(c *object.Commit) *domain.Commit {
	return &domain.Commit{
		Hash:      c.Hash.String(),
		ShortHash: c.Hash.String()[:7],
		Message:   c.Message,
		Author:    c.Author.Name,
		Email:     c.Author.Email,
		When:      c.Author.When,
	}
}

func isBinary(content []byte) bool {
	n := len(content)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(content[:n], 0) >= 0
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, gogit.ErrRepositoryNotExists),
		errors.Is(err, plumbing.ErrReferenceNotFound),
		errors.Is(err, plumbing.ErrObjectNotFound),
		errors.Is(err, object.ErrFileNotFound),
		errors.Is(err, object.ErrDirectoryNotFound),
		errors.Is(err, object.ErrEntryNotFound):
		return domain.ErrNotFound
	default:
		return err
	}
}
