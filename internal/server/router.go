package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"vk-intern/internal/models"
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
		return writeErr(r, http.StatusBadRequest, ErrNoLogPass)
	}
	defer r.Body.Close()

	if loginReq.Username == "" || loginReq.Password == "" {
		log.Error("Нет данных для входа")
		return writeErr(r, http.StatusBadRequest, ErrNoLogPass)
	}

	token, err := s.auth.Login(r.Context(), s.cfg.Secret,
		loginReq.Username, loginReq.Password, s.cfg.Server.Token,
	)
	if err != nil {
		log.Error("Ошибка авторизации пользователя",
			slog.String("error", err.Error()))
		return writeErr(r, http.StatusUnauthorized, ErrInvalidCred)
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
		log.Error("Не удалось преобразовать запрос в объект",
			slog.String("error", err.Error()))
		return writeErr(r, http.StatusBadRequest, ErrBadReq)
	}
	defer r.Body.Close()

	if len(writeReq.Data) == 0 {
		log.Error("Нет данных для записи")
		return writeErr(r, http.StatusBadRequest, ErrBadReq)
	}

	if err := s.storage.Write(r.Context(),
		s.cfg.Server.Timeout, writeReq.Data,
	); err != nil {
		log.Error("Не удалось записать данные",
			slog.String("error", err.Error()))
		return writeErr(r, http.StatusInternalServerError, err)
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
		log.Error("Не удалось преобразовать запрос в объект",
			slog.String("error", err.Error()))
		return writeErr(r, http.StatusBadRequest, ErrBadReq)
	}
	defer r.Body.Close()

	if len(readReq.Keys) == 0 {
		log.Error("Нет данных для чтения")
		return writeErr(r, http.StatusBadRequest, ErrBadReq)
	}

	data, err := s.storage.Read(r.Context(), s.cfg.Server.Timeout, readReq.Keys)
	if err != nil {
		log.Error("Не удалось прочитать данные",
			slog.String("error", err.Error()))
		return writeErr(r, http.StatusInternalServerError, err)
	}

	readResp := &models.ReadResponse{Data: data}
	return writeJSON(w, http.StatusOK, readResp)
}
