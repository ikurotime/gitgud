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

func (h *Handlers) issuesList(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}

	state := r.URL.Query().Get("state")
	var filter domain.IssueState
	switch state {
	case "closed":
		filter = domain.IssueClosed
	case "all":
		filter = ""
	default:
		state = "open"
		filter = domain.IssueOpen
	}

	issues, err := h.issues.List(r.Context(), repo.ID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	open, closed, err := h.issues.Counts(r.Context(), repo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, templates.IssuesList(currentUser(r.Context()), repo, state, issues, open, closed))
}

func (h *Handlers) newIssue(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	render(w, r, http.StatusOK, templates.IssueNew(currentUser(r.Context()), repo, "", "", ""))
}

func (h *Handlers) createIssue(w http.ResponseWriter, r *http.Request) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return
	}
	title := r.FormValue("title")
	body := r.FormValue("body")

	issue, err := h.issues.Open(r.Context(), repo, currentUser(r.Context()), title, body)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			msg := strings.TrimSuffix(err.Error(), ": "+domain.ErrValidation.Error())
			render(w, r, http.StatusBadRequest, templates.IssueNew(currentUser(r.Context()), repo, title, body, msg))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/issues/"+strconv.Itoa(issue.Number)), http.StatusSeeOther)
}

func (h *Handlers) issueDetail(w http.ResponseWriter, r *http.Request) {
	repo, issue := h.loadIssue(w, r)
	if issue == nil {
		return
	}
	comments, err := h.issues.Comments(r.Context(), issue.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	canModify := app.CanModifyIssue(repo, issue, currentUser(r.Context()))
	render(w, r, http.StatusOK, templates.IssueDetail(currentUser(r.Context()), repo, issue, comments, canModify))
}

func (h *Handlers) addIssueComment(w http.ResponseWriter, r *http.Request) {
	repo, issue := h.loadIssue(w, r)
	if issue == nil {
		return
	}
	if err := h.issues.Comment(r.Context(), issue, currentUser(r.Context()), r.FormValue("body")); err != nil {
		h.gitError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/issues/"+strconv.Itoa(issue.Number)), http.StatusSeeOther)
}

func (h *Handlers) closeIssue(w http.ResponseWriter, r *http.Request) {
	repo, issue := h.loadIssue(w, r)
	if issue == nil {
		return
	}
	if err := h.issues.Close(r.Context(), repo, issue, currentUser(r.Context())); err != nil {
		h.gitError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/issues/"+strconv.Itoa(issue.Number)), http.StatusSeeOther)
}

func (h *Handlers) reopenIssue(w http.ResponseWriter, r *http.Request) {
	repo, issue := h.loadIssue(w, r)
	if issue == nil {
		return
	}
	if err := h.issues.Reopen(r.Context(), repo, issue, currentUser(r.Context())); err != nil {
		h.gitError(w, r, err)
		return
	}
	http.Redirect(w, r, repoURL(repo, "/issues/"+strconv.Itoa(issue.Number)), http.StatusSeeOther)
}

func (h *Handlers) loadIssue(w http.ResponseWriter, r *http.Request) (*domain.Repository, *domain.Issue) {
	repo := h.viewRepo(w, r)
	if repo == nil {
		return nil, nil
	}
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		h.notFound(w, r)
		return nil, nil
	}
	issue, _, err := h.issues.Get(r.Context(), repo.ID, number)
	if err != nil {
		h.gitError(w, r, err)
		return nil, nil
	}
	return repo, issue
}

func repoURL(repo *domain.Repository, sub string) string {
	return "/" + repo.OwnerName + "/" + repo.Name + sub
}
