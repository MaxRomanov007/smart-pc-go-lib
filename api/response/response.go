package response

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response/pagination"
	"github.com/go-playground/validator/v10"
)

type Response[T any] struct {
	Status     string                 `json:"status"`
	Error      string                 `json:"error,omitempty"`
	Data       *T                     `json:"data,omitempty"`
	Pagination *pagination.Pagination `json:"pagination,omitempty"`
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

func ValidationError(errs validator.ValidationErrors) *Response[types.Nil] {
	var errMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(
				errMessages,
				fmt.Sprintf("field %q is a required field", err.Field()),
			)
		case "url":
			errMessages = append(
				errMessages,
				fmt.Sprintf("field %q is not a valid URL", err.Field()),
			)
		case "max":
			errMessages = append(
				errMessages,
				fmt.Sprintf("field %q must not exceed %s", err.Field(), err.Param()),
			)
		case "min":
			errMessages = append(
				errMessages,
				fmt.Sprintf("field %q must be at least %s", err.Field(), err.Param()),
			)
		default:
			errMessages = append(errMessages, fmt.Sprintf("field %q is not valid", err.Field()))
		}
	}

	return Error(StatusBadRequest, strings.Join(errMessages, ", "))
}

func (r *Response[T]) WithPagination(p *pagination.Pagination) *Response[T] {
	r.Pagination = p
	return r
}
