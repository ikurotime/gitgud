package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"gitgud/internal/domain"
)

func (h *Handlers) gitHTTP(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repoName := strings.TrimSuffix(chi.URLParam(r, "repo"), ".git")

	isPush := strings.HasSuffix(r.URL.Path, "git-receive-pack") ||
		r.URL.Query().Get("service") == "git-receive-pack"

	user := h.basicAuthUser(r)

	allow, needAuth, remoteUser, err := h.gitAccess.Authorize(r.Context(), owner, repoName, isPush, user)
	if needAuth {
		requireBasicAuth(w)
		return
	}
	if err != nil {
		if errors.Is(err, domain.ErrPermission) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.NotFound(w, r)
		return
	}
	if !allow {
		http.NotFound(w, r)
		return
	}

	h.gitBackend.Handler(remoteUser).ServeHTTP(w, r)
}

func (h *Handlers) basicAuthUser(r *http.Request) *domain.User {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil
	}
	u, err := h.users.Authenticate(r.Context(), username, password)
	if err != nil {
		return nil
	}
	return u
}

func requireBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="gitgud"`)
	http.Error(w, "authentication required", http.StatusUnauthorized)
}
