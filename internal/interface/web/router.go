package web

import (
	"embed"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gitgud/internal/infra/config"
)

var staticFS embed.FS

func NewRouter(cfg config.Config, h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(h.sm.LoadAndSave)
	r.Use(csrf)
	r.Use(h.withFlash)
	r.Use(h.withUser)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer((http.FS(staticFS)))))

	r.Get("/register", h.showRegister)
	r.Post("/register", h.doRegister)
	r.Get("/login", h.showLogin)
	r.Post("/login", h.doLogin)
	r.Post("/logout", h.doLogout)

	r.With(h.requireAuth).Get("/new", h.showNewRepo)
	r.With(h.requireAuth).Post("/new", h.createRepo)

	r.Get("/{owner}/{repo}/info/refs", h.gitHTTP)
	r.Post("/{owner}/{repo}/git-upload-pack", h.gitHTTP)
	r.Post("/{owner}/{repo}/git-receive-pack", h.gitHTTP)

	r.Get("/", h.dashboard)
	r.Get("/{owner}", h.profile)
	r.Get("/{owner}/{repo}", h.repoTreeRoot)
	r.Get("/{owner}/{repo}/tree/{ref}", h.repoTree)
	r.Get("/{owner}/{repo}/tree/{ref}/*", h.repoTree)
	r.Get("/{owner}/{repo}/blob/{ref}/*", h.repoBlob)
	r.Get("/{owner}/{repo}/raw/{ref}/*", h.repoRaw)
	r.Get("/{owner}/{repo}/commits/{ref}", h.repoCommits)
	r.Get("/{owner}/{repo}/commit/{hash}", h.repoCommit)

	r.Get("/{owner}/{repo}/issues", h.issuesList)
	r.With(h.requireAuth).Get("/{owner}/{repo}/issues/new", h.newIssue)
	r.With(h.requireAuth).Post("/{owner}/{repo}/issues", h.createIssue)
	r.Get("/{owner}/{repo}/issues/{number}", h.issueDetail)
	r.With(h.requireAuth).Post("/{owner}/{repo}/issues/{number}/comments", h.addIssueComment)
	r.With(h.requireAuth).Post("/{owner}/{repo}/issues/{number}/close", h.closeIssue)
	r.With(h.requireAuth).Post("/{owner}/{repo}/issues/{number}/reopen", h.reopenIssue)

	r.Get("/{owner}/{repo}/pulls", h.pullsList)
	r.Get("/{owner}/{repo}/compare", h.comparePull)
	r.With(h.requireAuth).Get("/{owner}/{repo}/pulls/new", h.comparePull)
	r.With(h.requireAuth).Post("/{owner}/{repo}/pulls", h.createPull)
	r.Get("/{owner}/{repo}/pulls/{number}", h.pullDetail)
	r.With(h.requireAuth).Post("/{owner}/{repo}/pulls/{number}/comments", h.addPullComment)
	r.With(h.requireAuth).Post("/{owner}/{repo}/pulls/{number}/merge", h.mergePull)
	r.With(h.requireAuth).Post("/{owner}/{repo}/pulls/{number}/close", h.closePull)

	return r
}

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
