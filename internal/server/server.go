package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"vk-intern/internal/config"
)

type Storage interface {
	GetUser(ctx context.Context, username string, passHash []byte) (string, error)

	Write(ctx context.Context, data map[string]any) error
	Read(ctx context.Context, keys []string) (map[string]any, error)
}

type Server struct {
	cfg *config.Config
	log *slog.Logger

	storage Storage
}

func New(cfg *config.Config, log *slog.Logger, storage Storage) *Server {
	return &Server{
		cfg: cfg,
		log: log,

		storage: storage,
	}
}

func (s *Server) Run() error {
	const op = "server.Run"
	log := s.log.With("op", op)

	// addr := s.cfg.Server.Host + ":" + s.cfg.Server.Port
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	router := s.newRouter()
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Info("Starting server")
	if err := server.ListenAndServe(); err != nil {
		log.Error("Error occured while server was running")
		panic(err)
	}
	log.Info("Server is running")

	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	// <-quit

	// make const\env var
	ctx, shutdown := context.WithTimeout(context.Background(), s.cfg.Server.Timeout)
	defer shutdown()

	return server.Shutdown(ctx)
}
