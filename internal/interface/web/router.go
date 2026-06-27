package web

import (
	"embed"
	"gitgud/internal/infra/config"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var templatesFS embed.FS

var staticFS embed.FS

func NewRouter(cfg config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger, middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer((http.FS(staticFS)))))
	r.Get("/", homeHandler)
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to GitGud!"))
}
