package main

import (
	"gitgud/internal/infra/config"
	"gitgud/internal/infra/persistence/sqlite"
	"gitgud/internal/interface/web"
	"log"
	"net/http"
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

	handler := web.NewRouter(cfg)

	log.Printf("listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
