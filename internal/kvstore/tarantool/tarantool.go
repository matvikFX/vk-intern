package tarantool

import (
	"context"
	"fmt"
	"log/slog"

	"vk-intern/internal/config"
	"vk-intern/internal/models"

	"github.com/tarantool/go-tarantool/v2"
)

type Tarantool struct {
	cfg *config.TarantoolConfig
	log *slog.Logger

	conn *tarantool.Connection
}

func New(cfg *config.TarantoolConfig, logger *slog.Logger) (*Tarantool, error) {
	const op = "tarantool.New"
	log := logger.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address: cfg.Host,
		User:    "admin",
	}
	opts := tarantool.Opts{
		Timeout: cfg.Timeout,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Error("Не удалось подключиться к Tarantool", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := conn.Do(&tarantool.PingRequest{}).Get(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Tarantool{
		cfg: cfg,
		log: log,

		conn: conn,
	}, nil
}

func (t *Tarantool) Stop() {
	t.conn.Close()
}

func (t *Tarantool) GetUser(ctx context.Context, username string) (*models.User, error) {
	const op = "tarantool.GetUser"
	log := t.log.With(slog.String("op", op))

	log.Info("GetUser")
	return nil, fmt.Errorf("%s: %w", op, ErrImplementMe)
}

func (t *Tarantool) Write(ctx context.Context, data models.Data) error {
	const op = "tarantool.Write"
	log := t.log.With(slog.String("op", op))

	log.Info("Write")
	return fmt.Errorf("%s: %w", op, ErrImplementMe)
}

func (t *Tarantool) Read(ctx context.Context, keys []string) (models.Data, error) {
	const op = "tarantool.Read"
	log := t.log.With(slog.String("op", op))

	log.Info("Read")
	return nil, fmt.Errorf("%s: %w", op, ErrImplementMe)
}
