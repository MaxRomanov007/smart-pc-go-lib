package response

import (
	"go/types"
)

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
	StatusOK            = "ok"
	StatusBadRequest    = "bad-request"
	StatusNotFound      = "not-found"
	StatusUnauthorized  = "unauthorized"
	StatusForbidden     = "forbidden"
	StatusInternalError = "internal-error"
)

func OK[T any](data *T) *Response[T] {
	return New(StatusOK, data, "")
}

func Error(status, msg string) *Response[types.Nil] {
	return New[types.Nil](status, nil, msg)
}

func ErrorWithMessagePrefix(status, prefix, msg string) *Response[types.Nil] {
	errMsg := prefix
	if msg != "" {
		errMsg += ": " + msg
	}
	return Error(status, errMsg)
}

func BadRequest(msg string) *Response[types.Nil] {
	return Error(StatusBadRequest, msg)
}

func Unauthorized(msg string) *Response[types.Nil] {
	return ErrorWithMessagePrefix(StatusUnauthorized, "Unauthorized", msg)
}

func Forbidden(msg string) *Response[types.Nil] {
	return ErrorWithMessagePrefix(StatusForbidden, "Forbidden", msg)
}

func NotFound(msg string) *Response[types.Nil] {
	return ErrorWithMessagePrefix(StatusNotFound, "Not Found", msg)
}

func InternalError() *Response[types.Nil] {
	return Error(StatusInternalError, "Internal error")
}
