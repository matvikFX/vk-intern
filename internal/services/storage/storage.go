package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"vk-intern/internal/models"
)

type KVStore interface {
	Write(ctx context.Context, data models.Data) error
	Read(ctx context.Context, keys []string) (models.Data, error)
}

type Storage struct {
	log     *slog.Logger
	kvStore KVStore
}

func New(log *slog.Logger, kvStore KVStore) *Storage {
	return &Storage{
		log:     log,
		kvStore: kvStore,
	}
}

func (s *Storage) Write(ctx context.Context, timeout time.Duration, data models.Data) error {
	const op = "service.Write"
	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Info("Запись в базу данных")
	if err := s.kvStore.Write(ctx, data); err != nil {
		log.Error("Ошибка при записи в базу данных", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Запись прошла успешно")

	return nil
}

func (s *Storage) Read(ctx context.Context, timeout time.Duration, keys []string) (models.Data, error) {
	const op = "service.Read"
	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Info("Чтение базы данных")
	data, err := s.kvStore.Read(ctx, keys)
	if err != nil {
		log.Error("Ошибка при чтении из базы данных", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Чтение прошло успешно")

	return data, nil
}
