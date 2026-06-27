package main

import (
	"log"
	"net/http"

	"gitgud/internal/app"
	"gitgud/internal/infra/config"
	"gitgud/internal/infra/persistence/sqlite"
	"gitgud/internal/infra/security"
	"gitgud/internal/infra/session"
	"gitgud/internal/interface/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlite.Open(cfg.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hasher := security.NewBcryptHasher()
	userRepo := sqlite.NewUserRepo(db)
	userService := app.NewUserService(userRepo, hasher)
	sm := session.NewSessionManager(db)
	handlers := web.NewHandlers(userService, sm)

	handler := web.NewRouter(cfg, handlers)

	log.Printf("listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
