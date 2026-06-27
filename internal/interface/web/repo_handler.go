package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"gitgud/internal/app"
	"gitgud/internal/domain"
	"gitgud/internal/interface/web/templates"
)

func (h *Handlers) dashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	if user == nil {
		render(w, r, http.StatusOK, templates.Home(nil))
		return
	}

	repos, err := h.repos.ListByOwner(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, templates.Dashboard(user, repos))
}

func (h *Handlers) showNewRepo(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, templates.NewRepo(currentUser(r.Context()), "", "", false, ""))
}

func (h *Handlers) createRepo(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	name := r.FormValue("name")
	description := r.FormValue("description")
	private := r.FormValue("private") != ""

	repo, err := h.repos.CreateRepo(r.Context(), user, name, description, private)
	if err != nil {
		status := http.StatusBadRequest
		msg := "could not create repository"
		switch {
		case errors.Is(err, domain.ErrConflict):
			status = http.StatusConflict
			msg = "a repository with that name already exists"
		case errors.Is(err, domain.ErrValidation):
			msg = strings.TrimSuffix(err.Error(), ": "+domain.ErrValidation.Error())
		}
		render(w, r, status, templates.NewRepo(user, name, description, private, msg))
		return
	}

	http.Redirect(w, r, "/"+repo.OwnerName+"/"+repo.Name, http.StatusSeeOther)
}

func (h *Handlers) profile(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r.Context())
	ownerName := chi.URLParam(r, "owner")

	owner, err := h.users.ByUsername(r.Context(), ownerName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	repos, err := h.repos.ListByOwner(r.Context(), owner.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	visible := repos[:0]
	for _, repo := range repos {
		if app.CanView(repo, viewer) == nil {
			visible = append(visible, repo)
		}
	}
	render(w, r, http.StatusOK, templates.Profile(viewer, owner.Username, visible))
}

func (h *Handlers) notFound(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusNotFound, templates.NotFound(currentUser(r.Context())))
}
