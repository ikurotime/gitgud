package web

import (
	"embed"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gitgud/internal/infra/config"
	"gitgud/internal/interface/web/templates"
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

	r.Get("/", homeHandler)

	r.Get("/register", h.showRegister)
	r.Post("/register", h.doRegister)
	r.Get("/login", h.showLogin)
	r.Post("/login", h.doLogin)
	r.Post("/logout", h.doLogout)

	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, templates.Home(currentUser(r.Context())))
}

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
