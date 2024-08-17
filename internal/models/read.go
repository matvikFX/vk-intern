package models

type ReadRequest struct {
	Keys []string `json:"keys"`
}

type ReadResponse struct {
	Data map[string]any `json:"data"`
}
