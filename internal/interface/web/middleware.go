package web

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"

	"gitgud/internal/app"
	"gitgud/internal/domain"
	"gitgud/internal/infra/git"
	"gitgud/internal/interface/web/templates"
)

const sessionUserIDKey = "user_id"

type ctxKey int

const userCtxKey ctxKey = iota

type Handlers struct {
	users      *app.UserService
	repos      *app.RepoService
	browse     *app.BrowseService
	issues     *app.IssueService
	pulls      *app.PullService
	gitAccess  *app.GitAccessService
	gitBackend *git.Backend
	sm         *scs.SessionManager
}

func NewHandlers(users *app.UserService, repos *app.RepoService, browse *app.BrowseService, issues *app.IssueService, pulls *app.PullService, gitAccess *app.GitAccessService, gitBackend *git.Backend, sm *scs.SessionManager) *Handlers {
	return &Handlers{users: users, repos: repos, browse: browse, issues: issues, pulls: pulls, gitAccess: gitAccess, gitBackend: gitBackend, sm: sm}
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

func (h *Handlers) withFlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if msg := h.sm.PopString(r.Context(), "flash"); msg != "" {
			r = r.WithContext(templates.WithFlash(r.Context(), msg))
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) flash(r *http.Request, msg string) {
	h.sm.Put(r.Context(), "flash", msg)
}

func csrf(next http.Handler) http.Handler {
	s := nosurf.New(injectCSRFToken(next))
	s.ExemptRegexp(`/(git-upload-pack|git-receive-pack)$`)
	s.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "CSRF token invalid", http.StatusBadRequest)
	}))
	return s
}

func injectCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(templates.WithCSRF(r.Context(), nosurf.Token(r))))
	})
}
