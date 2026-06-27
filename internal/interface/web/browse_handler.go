package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"gitgud/internal/domain"
	"gitgud/internal/interface/web/presenter"
	"gitgud/internal/interface/web/templates"
)

func (h *Handlers) repoTreeRoot(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	h.renderTree(w, r, repo, "", "")
}

func (h *Handlers) repoTree(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	h.renderTree(w, r, repo, chi.URLParam(r, "ref"), chi.URLParam(r, "*"))
}

func (h *Handlers) renderTree(w http.ResponseWriter, r *http.Request, repo *domain.Repository, ref, path string) {
	ctx := r.Context()

	empty, err := h.browse.IsEmpty(ctx, repo)
	if err != nil {
		h.gitError(w, r, err)
		return
	}
	if empty {
		render(w, r, http.StatusOK, templates.RepoHome(currentUser(ctx), repo))
		return
	}

	if ref == "" {
		ref = repo.DefaultBranch
	}

	branches, err := h.browse.Branches(ctx, repo)
	if err != nil {
		h.gitError(w, r, err)
		return
	}
	entries, err := h.browse.Tree(ctx, repo, ref, path)
	if err != nil {
		h.gitError(w, r, err)
		return
	}

	readme := ""
	for _, e := range entries {
		if !e.IsDir && strings.EqualFold(e.Name, "README.md") {
			if blob, err := h.browse.Blob(ctx, repo, ref, e.Path); err == nil && !blob.IsBinary {
				readme = presenter.RenderMarkdown(blob.Content)
			}
			break
		}
	}

	render(w, r, http.StatusOK, templates.BrowseTree(currentUser(ctx), repo, ref, path, branches, entries, readme))
}

func (h *Handlers) repoBlob(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	ref := chi.URLParam(r, "ref")
	path := chi.URLParam(r, "*")

	blob, err := h.browse.Blob(r.Context(), repo, ref, path)
	if err != nil {
		h.gitError(w, r, err)
		return
	}
	if ref == "" {
		ref = repo.DefaultBranch
	}

	highlighted := ""
	if !blob.IsBinary {
		highlighted = presenter.Highlight(string(blob.Content), blob.Path)
	}
	render(w, r, http.StatusOK, templates.BrowseBlob(currentUser(r.Context()), repo, ref, blob, highlighted))
}

func (h *Handlers) repoRaw(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	blob, err := h.browse.Blob(r.Context(), repo, chi.URLParam(r, "ref"), chi.URLParam(r, "*"))
	if err != nil {
		h.gitError(w, r, err)
		return
	}

	ctype := http.DetectContentType(blob.Content)
	if !blob.IsBinary {
		ctype = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ctype)
	w.Write(blob.Content)
}

func (h *Handlers) repoCommits(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	ref := chi.URLParam(r, "ref")

	commits, err := h.browse.Log(r.Context(), repo, ref, 50, 0)
	if err != nil {
		h.gitError(w, r, err)
		return
	}
	if ref == "" {
		ref = repo.DefaultBranch
	}
	render(w, r, http.StatusOK, templates.Commits(currentUser(r.Context()), repo, ref, commits))
}

func (h *Handlers) repoCommit(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	commit, diffs, err := h.browse.CommitDiff(r.Context(), repo, chi.URLParam(r, "hash"))
	if err != nil {
		h.gitError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, templates.CommitDetail(currentUser(r.Context()), repo, commit, diffs))
}

func (h *Handlers) viewRepo(w http.ResponseWriter, r *http.Request) *domain.Repository {
	repo, err := h.browse.Repo(r.Context(), chi.URLParam(r, "owner"), chi.URLParam(r, "repo"), currentUser(r.Context()))
	if err != nil {
		h.notFound(w, r)
		return nil
	}
	return repo
}

func (h *Handlers) gitError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		h.notFound(w, r)
	case errors.Is(err, domain.ErrPermission):
		http.Error(w, "forbidden", http.StatusForbidden)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
