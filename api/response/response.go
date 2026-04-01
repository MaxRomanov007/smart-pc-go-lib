package response

import "go/types"

type Response[T any] struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   *T     `json:"data,omitempty"`
}

func New[T any](status string, data *T, error string) *Response[T] {
	return &Response[T]{
		Status: status,
		Error:  error,
		Data:   data,
	}
}

const (
	StatusOK    = "ok"
	StatusError = "error"
)

func OK[T any](data *T) Response[T] {
	return Response[T]{
		Status: StatusOK,
		Data:   data,
	}
}

func Error(msg string) Response[types.Nil] {
	return Response[types.Nil]{
		Status: StatusError,
		Error:  msg,
	}
}
