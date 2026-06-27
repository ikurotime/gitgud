package web

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"

	"gitgud/internal/app"
	"gitgud/internal/domain"
	"gitgud/internal/infra/git"
)

const sessionUserIDKey = "user_id"

type ctxKey int

const userCtxKey ctxKey = iota

type Handlers struct {
	users      *app.UserService
	repos      *app.RepoService
	browse     *app.BrowseService
	gitAccess  *app.GitAccessService
	gitBackend *git.Backend
	sm         *scs.SessionManager
}

func NewHandlers(users *app.UserService, repos *app.RepoService, browse *app.BrowseService, gitAccess *app.GitAccessService, gitBackend *git.Backend, sm *scs.SessionManager) *Handlers {
	return &Handlers{users: users, repos: repos, browse: browse, gitAccess: gitAccess, gitBackend: gitBackend, sm: sm}
}

func (h *Handlers) withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := h.sm.GetInt64(r.Context(), sessionUserIDKey); id != 0 {
			if u, err := h.users.ByID(r.Context(), strconv.FormatInt(id, 10)); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, u))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r.Context()) == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userCtxKey).(*domain.User)
	return u
}
