package main

import (
	"log/slog"
	"os"

	"vk-intern/internal/config"
	"vk-intern/internal/kvstore/tarantool"
	"vk-intern/internal/server"
	"vk-intern/internal/services/auth"
	"vk-intern/internal/services/storage"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := newLogger(cfg.Env)

	// kvStore
	tarantool, err := tarantool.New(&cfg.Tarantool, log)
	if err != nil {
		log.Error("Ошибка подключения Tarantool", slog.String("error", err.Error()))
		panic(err)
	}
	defer tarantool.Stop()

	// services
	auth := auth.New(log, tarantool)
	storage := storage.New(log, tarantool)

	// server
	server := server.New(cfg, log, auth, storage)

	if err := server.Run(); err != nil {
		log.Error("Ошибка при работе сервера", slog.String("error", err.Error()))
		panic(err)
	}
}

func newLogger(env string) *slog.Logger {
	if env == envLocal {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
