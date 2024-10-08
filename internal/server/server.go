package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vk-intern/internal/config"
	"vk-intern/internal/models"
)

var (
	ErrNoLogPass   = errors.New("Поля логина или пароля не должны быть пустые")
	ErrBadReq      = errors.New("Некорректные данные запроса")
	ErrUnauth      = errors.New("Пользователь не авторизован")
	ErrInvalidCred = errors.New("Неправильный логин или пароль")
)

type Auth interface {
	Login(ctx context.Context, secret, username, password string, duration time.Duration) (string, error)
	FindUser(ctx context.Context, username string) error
}

type Storage interface {
	Write(ctx context.Context, timeout time.Duration, data models.Data) error
	Read(ctx context.Context, timeout time.Duration, keys []string) (models.Data, error)
}

type Server struct {
	cfg *config.Config
	log *slog.Logger

	auth    Auth
	storage Storage
}

func New(cfg *config.Config, log *slog.Logger, auth Auth, storage Storage) *Server {
	return &Server{
		cfg: cfg,
		log: log,

		auth:    auth,
		storage: storage,
	}
}

func (s *Server) Run() error {
	const op = "server.Run"
	log := s.log.With(slog.String("op", op))

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	router := s.newRouter()
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Info("Запуск сервера")
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error("Произошла ошибка во время работы сервера")
			panic(err)
		}
	}()
	log.Info(fmt.Sprintf("Сервер слушает порт %d", s.cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Завершение работы сервера")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
