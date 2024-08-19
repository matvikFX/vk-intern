package models

type Data map[string]any

// api/read
type ReadRequest struct {
	Keys []string `json:"keys"`
}

type ReadResponse struct {
	Data Data `json:"data"`
}

// api/write
type WriteRequest struct {
	Data Data `json:"data"`
}

type WriteResponse struct {
	Status string `json:"status"`
}

// tarantool obj
type Pair struct {
	Key   string `msgpack:"key"`
	Value any    `msgpack:"value"`
}
