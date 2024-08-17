package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"vk-intern/internal/config"
	"vk-intern/internal/server"
	"vk-intern/internal/tarantool"
)

const (
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()
	log := newLogger(cfg.Env)

	log.Info("Setting Tarantool")
	tarantool := tarantool.New()

	log.Info("Setting server")
	server := server.New(cfg, log, tarantool)

	go server.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
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
