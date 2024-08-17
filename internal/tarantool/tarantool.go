package tarantool

import (
	"context"
	"errors"
)

var ErrImplementMe = errors.New("Implement me")

type Tarantool struct{}

func New() *Tarantool {
	return &Tarantool{}
}

func (t *Tarantool) GetUser(ctx context.Context, username string, passHash []byte) (string, error) {
	return "", ErrImplementMe
}

func (t *Tarantool) Write(ctx context.Context, data map[string]any) error {
	return ErrImplementMe
}

func (t *Tarantool) Read(ctx context.Context, keys []string) (map[string]any, error) {
	return nil, ErrImplementMe
}
