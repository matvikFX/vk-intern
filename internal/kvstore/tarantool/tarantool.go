package tarantool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"vk-intern/internal/config"
	"vk-intern/internal/kvstore"
	"vk-intern/internal/models"

	"github.com/tarantool/go-tarantool/v2"
)

var (
	ErrChanIsClosed = errors.New("Канал закрыт")
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

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	dialer := tarantool.NetDialer{
		Address:  addr,
		User:     cfg.User,
		Password: cfg.Pass,
	}
	opts := tarantool.Opts{
		Timeout: cfg.Timeout,
	}

	log.Info("Подключение к Tarantool", slog.String("addr", addr))
	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Error("Не удалось подключиться к Tarantool", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Проверка соединения с Tarantool")
	if _, err := conn.Do(tarantool.NewPingRequest()).Get(); err != nil {
		log.Error("Не удалось получить ответ от Tarantool", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Подключение к Tarantool прошло успешно")
	return &Tarantool{
		cfg: cfg,
		log: logger,

		conn: conn,
	}, nil
}

func (t *Tarantool) Stop() {
	t.conn.CloseGraceful()
}

func (t *Tarantool) GetUser(ctx context.Context, username string) (*models.User, error) {
	const op = "tarantool.GetUser"
	log := t.log.With(slog.String("op", op))

	log.Info("Получение пользователя")
	req := tarantool.NewSelectRequest("kv_users").
		Context(ctx).
		Index("primary").
		Limit(1).
		Iterator(tarantool.IterEq).
		Key(tarantool.StringKey{S: username})

	user := []*models.User{}
	if err := t.conn.Do(req).GetTyped(&user); err != nil {
		log.Error("Не удалось получить пользователя из БД", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(user) == 0 {
		log.Error("Пользователь не найден")
		return nil, fmt.Errorf("%s: %w", op, kvstore.ErrUserNotFound)
	}

	log.Info("Ответ получен", slog.Any("User", user[0]))
	return user[0], nil
}

func (t *Tarantool) Write(ctx context.Context, data models.Data) error {
	const op = "tarantool.Write"
	log := t.log.With(slog.String("op", op))

	errCh := make(chan error, 1)
	wg := sync.WaitGroup{}

	// Запись пар в канал
	for key, value := range data {
		wg.Add(1)
		go func(errCount chan<- error) {
			defer wg.Done()

			log.Info("Запись в БД", slog.String("key", key), slog.Any("value", value))
			req := tarantool.NewReplaceRequest("kv_storage").
				Context(ctx).
				Tuple([]interface{}{key, value})

			data, err := t.conn.Do(req).Get()
			if err != nil {
				errCh <- err
				close(errCh)

				// прерывает горутину
				log.Error("Не удалось записать данные в БД", slog.String("error", err.Error()))
				return
			}

			log.Info("Данные записаны в БД", slog.Any("data", data))
		}(errCh)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Если ошибок нет, значит все прошло успешно
	log.Info("Данные записаны в БД")
	return nil
}

func (t *Tarantool) Read(ctx context.Context, keys []string) (models.Data, error) {
	const op = "tarantool.Read"
	log := t.log.With(slog.String("op", op))

	wg := &sync.WaitGroup{}
	mu := sync.Mutex{}

	errCh := make(chan error, 1)
	data := make(models.Data)

	for _, key := range keys {
		wg.Add(1)
		go func(key string, errCh chan<- error) {
			defer wg.Done()

			req := tarantool.NewSelectRequest("kv_storage").
				Context(ctx).
				Index("primary").
				Limit(1).
				Iterator(tarantool.IterEq).
				Key(tarantool.StringKey{S: key})

			var pair []*models.Pair
			if err := t.conn.Do(req).GetTyped(&pair); err != nil {
				errCh <- err
				close(errCh)

				// прерывает горутину
				log.Error("Не удалось прочитать данные из БД", slog.String("error", err.Error()))
				return
			}

			if len(pair) == 0 {
				log.Error("Запись не найдена", slog.String("key", key))

				mu.Lock()
				data[key] = nil
				mu.Unlock()

				// прерывает горутину
				return
			}

			mu.Lock()
			data[key] = pair[0].Value
			mu.Unlock()
		}(key, errCh)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Данные из БД получены")
	return data, nil
}
