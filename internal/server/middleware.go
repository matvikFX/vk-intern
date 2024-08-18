package server

import (
	"log/slog"
	"net/http"
	"strings"

	"vk-intern/internal/jwt"
)

func (s *Server) withAuth(f handlerFunc) handlerFunc {
	const op = "server.withLogin"
	log := s.log.With(slog.String("op", op))

	return func(w http.ResponseWriter, r *http.Request) error {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			log.Error("Нет заголовка авторизации")
			return writeJSON(w, http.StatusUnauthorized, ErrUnauth)
		}

		token := strings.Split(auth, " ")[1]
		if token == "" {
			log.Error("Нет токена авторизации")
			return writeJSON(w, http.StatusUnauthorized, ErrUnauth)
		}

		log.Info("Проверка токена")
		username, err := jwt.ParseJWT(token, s.cfg.Secret)
		if err != nil {
			log.Error("Ошибка проверки токена", slog.String("error", err.Error()))
			return writeJSON(w, http.StatusUnauthorized, ErrUnauth)
		}
		log.Info("Токен проверен", slog.String("username", username))

		if err := s.auth.FindUser(r.Context(), username); err != nil {
			log.Error("Пользователь не найден")
			return writeJSON(w, http.StatusUnauthorized, ErrUnauth)
		}

		log.Info("User authorized")
		return f(w, r)
	}
}
