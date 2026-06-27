package main

import (
	"log"
	"net/http"
	"time"

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
	gitReader := git.NewGoGitReader(cfg.ReposDir())
	repoRepo := sqlite.NewRepoRepo(db)
	repoService := app.NewRepoService(repoRepo, gitSvc)
	browseService := app.NewBrowseService(repoRepo, gitReader)
	issueRepo := sqlite.NewIssueRepo(db)
	issueService := app.NewIssueService(issueRepo)
	prRepo := sqlite.NewPRRepo(db)
	pullService := app.NewPullService(prRepo, gitReader, gitSvc)
	gitAccess := app.NewGitAccessService(repoRepo)

	gitBackend, err := git.NewBackend(cfg.ReposDir())
	if err != nil {
		log.Fatal(err)
	}

	sm := session.NewSessionManager(db)
	handlers := web.NewHandlers(userService, repoService, browseService, issueService, pullService, gitAccess, gitBackend, sm)

	handler := web.NewRouter(cfg, handlers)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("listening on %s", cfg.Addr)
	log.Fatal(srv.ListenAndServe())
}
