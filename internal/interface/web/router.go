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

	r.Get("/", h.dashboard)
	r.Get("/{owner}", h.profile)
	r.Get("/{owner}/{repo}", h.repoHome)

	return r
}

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
