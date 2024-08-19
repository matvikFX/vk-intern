package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type User struct {
	Username string `msgpack:"username"`
	Password string `msgpack:"password"`
}
