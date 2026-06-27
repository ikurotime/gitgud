package main

import (
	"log"
	"net/http"

	"gitgud/internal/app"
	"gitgud/internal/infra/config"
	"gitgud/internal/infra/git"
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

	gitSvc := git.NewCLIGit(cfg.ReposDir())
	repoRepo := sqlite.NewRepoRepo(db)
	repoService := app.NewRepoService(repoRepo, gitSvc)
	gitAccess := app.NewGitAccessService(repoRepo)

	gitBackend, err := git.NewBackend(cfg.ReposDir())
	if err != nil {
		log.Fatal(err)
	}

	sm := session.NewSessionManager(db)
	handlers := web.NewHandlers(userService, repoService, gitAccess, gitBackend, sm)

	handler := web.NewRouter(cfg, handlers)

	log.Printf("listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
