package apiserver

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"tz/internal/config"
	"tz/internal/store"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Start(cfg config.Config) error {
	db, err := newDb(cfg.StoragePath)
	if err != nil {
		return err
	}
	defer db.Close()
	log := setupLogger(cfg.Env)
	store := store.NewPGStore(db)
	srv := NewServer(log, store)
	return http.ListenAndServe(cfg.Host+":"+cfg.Port, srv)

}

func newDb(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return db, nil
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
