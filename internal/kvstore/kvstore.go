package kvstore

import (
	"errors"
)

var (
	ErrUserNotFound = errors.New("Пользователь не найден")
	ErrDataNotFound = errors.New("Данные по ключу не найдены")
	ErrKeyNotFound  = errors.New("Ключи не найдены")
)
