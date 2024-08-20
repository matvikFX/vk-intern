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

const (
	numWriter = 4
	numReader = 4
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

	wg := sync.WaitGroup{}

	pairCh := make(chan *models.Pair, numWriter)
	errCh := make(chan error, 1)

	// Запись пар в канал
	go func() {
		for k, v := range data {
			t := &models.Pair{Key: k, Value: v}
			pairCh <- t
		}
		close(pairCh)
	}()

	// Запуск рабочих горутин для обработки запросов
	for range numWriter {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.writer(ctx, pairCh, errCh)
		}()
	}

	// Ожидание выполнения всех рабочих горутин
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Ожидание ошибок
	for err := range errCh {
		log.Error("Не удалось записать данные",
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	// Если ошибок нет, значит все прошло успешно
	log.Info("Данные записаны в БД")
	return nil
}

func (t *Tarantool) writer(ctx context.Context,
	pairCh <-chan *models.Pair, errCh chan<- error,
) {
	const op = "tarantool.writer"
	log := t.log.With(slog.String("op", op))

	for {
		select {
		case <-ctx.Done():
			log.Error("Контекст завершен", slog.String("error", ctx.Err().Error()))
			errCh <- ctx.Err()
			return
		case pair, ok := <-pairCh:
			if !ok {
				log.Error("Канал закрыт")
				return
			}

			log.Info("Запись в БД", slog.String("key", pair.Key), slog.Any("value", pair.Value))
			req := tarantool.NewReplaceRequest("kv_storage").
				Context(ctx).
				Tuple([]interface{}{pair.Key, pair.Value})

			data, err := t.conn.Do(req).Get()
			if err != nil {
				log.Error("Не удалось записать данные в БД", slog.String("error", err.Error()))
				errCh <- err
				return
			}

			log.Info("Данные записаны в БД", slog.Any("data", data))
		}
	}
}

func (t *Tarantool) Read(ctx context.Context, keys []string) (models.Data, error) {
	const op = "tarantool.Read"
	log := t.log.With(slog.String("op", op))

	wg := sync.WaitGroup{}

	keyCh := make(chan string, len(keys))
	pairCh := make(chan *models.Pair, len(keys))
	errCh := make(chan error, 1)

	// Запись запросов в канал
	log.Info("Запись ключей в канал")
	go func() {
		for _, key := range keys {
			keyCh <- key
		}
		close(keyCh)
	}()

	// Запуск рабочих горутин для обработки запросов
	log.Info("Запуск читающих горутин")
	for range numReader {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.reader(ctx, keyCh, pairCh, errCh)
		}()
	}

	// Ожидание выполнения всех рабочих горутин
	go func() {
		wg.Wait()
		close(pairCh)
		close(errCh)
	}()

	// Ожидание ошибок
	for err := range errCh {
		log.Info("Не удалось прочитать данные",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Запись полученых данных
	// Если ошибок нет, значит можно передавать данные
	data := make(models.Data)

	for pair := range pairCh {
		log.Info("Получены данные",
			slog.String("key", pair.Key), slog.Any("value", pair.Value))
		data[pair.Key] = pair.Value
	}

	log.Info("Данные из БД получены")
	return data, nil
}

func (t *Tarantool) reader(ctx context.Context,
	keyCh <-chan string, pairCh chan<- *models.Pair, errCh chan<- error,
) {
	const op = "tarantool.reader"
	log := t.log.With(slog.String("op", op))

	for {
		select {
		case <-ctx.Done():
			log.Error("Контекст завершен", slog.String("error", ctx.Err().Error()))
			errCh <- ctx.Err()
			return
		case key, ok := <-keyCh:
			if !ok {
				log.Error("Канал закрыт")
				return
			}

			req := tarantool.NewSelectRequest("kv_storage").
				Context(ctx).
				Index("primary").
				Limit(1).
				Iterator(tarantool.IterEq).
				Key(tarantool.StringKey{S: key})

			var pair []*models.Pair
			if err := t.conn.Do(req).GetTyped(&pair); err != nil {
				log.Error("Не удалось прочитать данные из БД", slog.String("error", err.Error()))
				errCh <- err
				return
			}

			if len(pair) == 0 {
				log.Error("Запись не найдена")

				// Думаю, не стоит прерываться
				pairCh <- &models.Pair{Key: key, Value: nil}
				continue
			}

			pairCh <- pair[0]
		}
	}
}
