package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"vk-intern/internal/jwt"
	"vk-intern/internal/kvstore"
	"vk-intern/internal/models"
)

type KVStore interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
}

var ErrInvalidCred = errors.New("Неправильный логин или пароль")

type Auth struct {
	log     *slog.Logger
	kvStore KVStore
}

func New(log *slog.Logger, kvStore KVStore) *Auth {
	return &Auth{
		log:     log,
		kvStore: kvStore,
	}
}

func (a *Auth) FindUser(ctx context.Context, username string) error {
	const op = "service.Login"
	log := a.log.With(slog.String("op", op))

	log.Info("Проверка на существование пользователя")
	if _, err := a.kvStore.GetUser(ctx, username); err != nil {
		if errors.Is(err, kvstore.ErrUserNotFound) {
			log.Error("Пользователь не найден")
			return fmt.Errorf("%s: %w", op, kvstore.ErrUserNotFound)
		}

		log.Error("Ошибка при проверки пользователя", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Пользователь найден")
	return nil
}

func (a *Auth) Login(ctx context.Context, secret, username, password string) (string, error) {
	const op = "service.Login"
	log := a.log.With(slog.String("op", op))

	if username == "" {
		log.Error("Имя является пустой строкой")
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCred)
	}

	if password == "" {
		log.Error("Пароль является пустой строкой")
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCred)
	}

	log.Info("Проверка пользователя")
	user, err := a.kvStore.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, kvstore.ErrUserNotFound) {
			log.Error("Пользователя с таким именем не существует")
			return "", fmt.Errorf("%s: %w", op, kvstore.ErrUserNotFound)
		}

		log.Error("Ошибка при получении пользователя", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if user.Password != password {
		log.Error("Неправильный пароль", slog.String("password", user.Password))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCred)
	}

	log.Info("Создание токена")
	token, err := jwt.NewToken(username, secret)
	if err != nil {
		log.Error("Не удалось создать токен")
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Токен успешно создан")

	return token, nil
}
