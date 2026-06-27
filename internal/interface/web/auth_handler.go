package web

import (
	"errors"
	"net/http"
	"strings"

	"gitgud/internal/domain"
	"gitgud/internal/interface/web/templates"
)

func (h *Handlers) showRegister(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, templates.Register(currentUser(r.Context()), "", "", ""))
}

func (h *Handlers) doRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := h.users.Register(r.Context(), username, email, password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		render(w, r, status, templates.Register(currentUser(r.Context()), username, email, userMessage(err)))
		return
	}
	h.startSession(w, r, u)
}

func (h *Handlers) showLogin(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, templates.Login(currentUser(r.Context()), "", ""))
}

func (h *Handlers) doLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	u, err := h.users.Authenticate(r.Context(), username, password)
	if err != nil {
		render(w, r, http.StatusUnauthorized, templates.Login(currentUser(r.Context()), username, "invalid credentials"))
		return
	}
	h.startSession(w, r, u)
}

func (h *Handlers) doLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.sm.Destroy(r.Context()); err != nil {
		h.writeError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) startSession(w http.ResponseWriter, r *http.Request, u *domain.User) {
	h.sm.Put(r.Context(), sessionUserIDKey, u.ID)
	h.sm.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func userMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return "invalid credentials"
	case errors.Is(err, domain.ErrConflict):
		return "that username is already taken"
	case errors.Is(err, domain.ErrValidation):
		return strings.TrimSuffix(err.Error(), ": "+domain.ErrValidation.Error())
	default:
		return "something went wrong"
	}
}
