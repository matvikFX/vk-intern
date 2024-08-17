package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"net/http"

	"vk-intern/internal/models"
)

const (
	tiemout = 10 * time.Second
)

var (
	ErrBadReq = errors.New("Поля логина или пароля не должны быть пустые")
	ErrUnauth = errors.New("Пользователь не авторизован")

	ErrNoUsername = errors.New("Имя является пустой строкой")
	ErrNoPassword = errors.New("Пароль является пустой строкой")
)

func (s *Server) newRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/hello", func(w http.ResponseWriter, r *http.Request) {
		resp := []byte("Hello, World!\n")
		w.Write(resp)
	})

	router.HandleFunc("POST /api/login", handleFunc(s.login))
	// Нужен токен
	router.HandleFunc("POST /api/write", handleFunc(s.write))
	router.HandleFunc("POST /api/read", handleFunc(s.read))
	// Надо реализовать что-то подобное
	// router.HandleFunc("POST /api/write", handleFunc(withLogin(s.write)))
	// router.HandleFunc("POST /api/read", handleFunc(withLogin(s.read)))

	return router
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) error {
	const op = "server.login"
	log := s.log.With("op", op)

	loginReq := &models.LoginRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		log.Error("Не удалось преобразовать запроса в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	defer r.Body.Close()

	if loginReq.Username == "" {
		log.Error(ErrNoUsername.Error())
		return writeJSON(w, http.StatusBadRequest, ErrBadReq)
	}

	if loginReq.Password == "" {
		log.Error(ErrNoPassword.Error())
		return writeJSON(w, http.StatusBadRequest, ErrBadReq)
	}

	// Добавить JWT токен
	log.Info("Создание токена")
	token := loginReq.Username + loginReq.Password
	log.Info("Токен успешно создан")

	loginResp := models.LoginResponse{Token: token}
	return writeJSON(w, http.StatusOK, loginResp)
}

func (s *Server) write(w http.ResponseWriter, r *http.Request) error {
	const op = "server.write"
	log := s.log.With("op", op)

	// Обязательно добавить проверку авторизации

	writeReq := &models.WriteRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(writeReq); err != nil {
		log.Error("Не удалось преобразовать запрос в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), tiemout)
	defer cancel()

	log.Info("Запись в базу данных")
	if err := s.storage.Write(ctx, writeReq.Data); err != nil {
		// if errors.As(err, storage.Error) {
		//
		// }
		log.Error("Ошибка при записи в базу данных", slog.String("error", err.Error()))
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	log.Info("Запись прошла успешно")

	writeResp := &models.WriteResponse{Status: "success"}
	return writeJSON(w, http.StatusCreated, writeResp)
}

func (s *Server) read(w http.ResponseWriter, r *http.Request) error {
	const op = "server.write"
	log := s.log.With("op", op)

	// Обязательно добавить проверку авторизации

	readReq := &models.ReadRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(readReq); err != nil {
		log.Error("Не удалось преобразовать запрос в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), tiemout)
	defer cancel()

	log.Info("Запись в базу данных")
	data, err := s.storage.Read(ctx, readReq.Keys)
	if err != nil {
		// if errors.As(err, storage.Error) {
		//
		// }
		log.Error("Ошибка при записи в базу данных", slog.String("error", err.Error()))
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	log.Info("Запись прошла успешно")

	readResp := &models.ReadResponse{Data: data}
	return writeJSON(w, http.StatusCreated, readResp)
}
