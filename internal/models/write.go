package models

type WriteRequest struct {
	Data map[string]any `json:"data"`
}

type WriteResponse struct {
	Status string `json:"status"`
}
