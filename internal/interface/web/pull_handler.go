package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"

	"gitgud/internal/app"
	"gitgud/internal/domain"
	"gitgud/internal/interface/web/templates"
)

func (h *Handlers) pullsList(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}

	state := r.URL.Query().Get("state")
	var filter domain.PRState
	switch state {
	case "merged":
		filter = domain.PRMerged
	case "closed":
		filter = domain.PRClosed
	case "all":
		filter = ""
	default:
		state = "open"
		filter = domain.PROpen
	}

	prs, err := h.pulls.List(r.Context(), repo.ID, filter)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	open, merged, closed, err := h.pulls.Counts(r.Context(), repo.ID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, templates.PullsList(currentUser(r.Context()), repo, state, prs, open, merged, closed))
}

func (h *Handlers) comparePull(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}

	branches, err := h.browse.Branches(r.Context(), repo)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	base := r.URL.Query().Get("base")
	if base == "" {
		base = repo.DefaultBranch
	}
	head := r.URL.Query().Get("head")

	var cmp *domain.Comparison
	errMsg := ""
	if head != "" && head != base {
		cmp, err = h.pulls.Compare(r.Context(), repo, base, head)
		if err != nil {
			errMsg = "could not compare those branches"
		}
	}
	render(w, r, http.StatusOK, templates.Compare(currentUser(r.Context()), repo, base, head, branches, cmp, errMsg))
}

func (h *Handlers) createPull(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	title := r.FormValue("title")
	body := r.FormValue("body")
	base := r.FormValue("base")
	head := r.FormValue("head")

	pr, err := h.pulls.Open(r.Context(), repo, currentUser(r.Context()), title, body, base, head)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			branches, _ := h.browse.Branches(r.Context(), repo)
			cmp, _ := h.pulls.Compare(r.Context(), repo, base, head)
			msg := strings.TrimSuffix(err.Error(), ": "+domain.ErrValidation.Error())
			render(w, r, http.StatusBadRequest, templates.Compare(currentUser(r.Context()), repo, base, head, branches, cmp, msg))
			return
		}
		h.writeError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/pulls/"+strconv.Itoa(pr.Number)), http.StatusSeeOther)
}

func (h *Handlers) pullDetail(w http.ResponseWriter, r *http.Request) {
	repo, pr := h.loadPull(w, r)
	if pr == nil {
		return
	}
	h.renderPull(w, r, repo, pr, "", http.StatusOK)
}

func (h *Handlers) addPullComment(w http.ResponseWriter, r *http.Request) {
	repo, pr := h.loadPull(w, r)
	if pr == nil {
		return
	}
	if err := h.pulls.Comment(r.Context(), pr, currentUser(r.Context()), r.FormValue("body")); err != nil {
		h.writeError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/pulls/"+strconv.Itoa(pr.Number)), http.StatusSeeOther)
}

func (h *Handlers) mergePull(w http.ResponseWriter, r *http.Request) {
	repo, pr := h.loadPull(w, r)
	if pr == nil {
		return
	}
	err := h.pulls.Merge(r.Context(), repo, pr, currentUser(r.Context()))
	if err == nil {
		h.flash(r, "Pull request merged.")
		http.Redirect(w, r, repoURL(repo, "/pulls/"+strconv.Itoa(pr.Number)), http.StatusSeeOther)
		return
	}
	if errors.Is(err, domain.ErrConflict) {
		h.renderPull(w, r, repo, pr, "Cannot merge automatically — the branches conflict.", http.StatusConflict)
		return
	}
	h.writeError(w, r, err)
}

func (h *Handlers) closePull(w http.ResponseWriter, r *http.Request) {
	repo, pr := h.loadPull(w, r)
	if pr == nil {
		return
	}
	if err := h.pulls.Close(r.Context(), repo, pr, currentUser(r.Context())); err != nil {
		h.writeError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/pulls/"+strconv.Itoa(pr.Number)), http.StatusSeeOther)
}

func (h *Handlers) renderPull(w http.ResponseWriter, r *http.Request, repo *domain.Repository, pr *domain.PullRequest, errMsg string, status int) {
	cmp, _ := h.pulls.Compare(r.Context(), repo, pr.BaseBranch, pr.HeadBranch)
	comments, err := h.pulls.Comments(r.Context(), pr.ID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	canMerge := pr.State == domain.PROpen && app.CanMergePR(repo, currentUser(r.Context())) && cmp != nil && cmp.Mergeable
	render(w, r, status, templates.PullDetail(currentUser(r.Context()), repo, pr, cmp, comments, canMerge, errMsg))
}

func (h *Handlers) loadPull(w http.ResponseWriter, r *http.Request) (*domain.Repository, *domain.PullRequest) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return nil, nil
	}
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		h.notFound(w, r)
		return nil, nil
	}
	pr, err := h.pulls.Get(r.Context(), repo.ID, number)
	if err != nil {
		h.writeError(w, r, err)
		return nil, nil
	}
	return repo, pr
}
