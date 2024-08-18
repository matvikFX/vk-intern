package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"vk-intern/internal/models"
)

type Auth interface {
	Login(ctx context.Context, secret, username, password string) (string, error)
	FindUser(ctx context.Context, username string) error
}

type Storage interface {
	Write(ctx context.Context, timeout time.Duration, data models.Data) error
	Read(ctx context.Context, timeout time.Duration, keys []string) (models.Data, error)
}

var (
	ErrBadReq      = errors.New("Поля логина или пароля не должны быть пустые")
	ErrUnauth      = errors.New("Пользователь не авторизован")
	ErrInvalidCred = errors.New("Неправильный логин или пароль")
)

func (s *Server) newRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/login", handleFunc(s.login))
	router.HandleFunc("POST /api/write", handleFunc(s.withAuth(s.write)))
	router.HandleFunc("POST /api/read", handleFunc(s.withAuth(s.read)))

	return router
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) error {
	const op = "server.login"
	log := s.log.With(slog.String("op", op))

	loginReq := &models.LoginRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		log.Error("Не удалось преобразовать запроса в объект",
			slog.String("error", err.Error()))
		return writeJSON(w, http.StatusInternalServerError, ErrBadReq)
	}
	defer r.Body.Close()

	token, err := s.auth.Login(r.Context(), s.cfg.Secret, loginReq.Username, loginReq.Password)
	if err != nil {
		log.Error("Ошибка создания токена пользователя",
			slog.String("error", err.Error()))
		return writeJSON(w, http.StatusUnauthorized, ErrInvalidCred)
	}

	loginResp := &models.LoginResponse{Token: token}
	return writeJSON(w, http.StatusOK, loginResp)
}

func (s *Server) write(w http.ResponseWriter, r *http.Request) error {
	const op = "server.write"
	log := s.log.With(slog.String("op", op))

	writeReq := &models.WriteRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(writeReq); err != nil {
		log.Error("Не удалось преобразовать запрос в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	defer r.Body.Close()

	if err := s.storage.Write(r.Context(), s.cfg.Server.Timeout, writeReq.Data); err != nil {
		log.Error("")
		return writeJSON(w, http.StatusInternalServerError, err)
	}

	writeResp := &models.WriteResponse{Status: "success"}
	return writeJSON(w, http.StatusCreated, writeResp)
}

func (s *Server) read(w http.ResponseWriter, r *http.Request) error {
	const op = "server.read"
	log := s.log.With(slog.String("op", op))

	readReq := &models.ReadRequest{}
	log.Info("Преобразование запроса в объект")
	if err := json.NewDecoder(r.Body).Decode(readReq); err != nil {
		log.Error("Не удалось преобразовать запрос в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}
	defer r.Body.Close()

	data, err := s.storage.Read(r.Context(), s.cfg.Server.Timeout, readReq.Keys)
	if err != nil {
		log.Error("Не удалось преобразовать запрос в объект")
		return writeJSON(w, http.StatusInternalServerError, err)
	}

	readResp := &models.ReadResponse{Data: data}
	return writeJSON(w, http.StatusCreated, readResp)
}
