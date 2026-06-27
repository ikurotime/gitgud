package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr       string
	DataDir    string
	SessionKey string // random string
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "app.db")
}

func (c Config) ReposDir() string {
	return filepath.Join(c.DataDir, "repos")
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	dataDir, err := filepath.Abs(os.Getenv("GITGUD_DATA_DIR"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Addr:       os.Getenv("GITGUD_ADDR"),
		DataDir:    dataDir,
		SessionKey: os.Getenv("GITGUD_SESSION_KEY"),
	}, nil
}
