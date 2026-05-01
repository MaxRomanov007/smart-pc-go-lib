package reqmw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ctxKey string

const requestKey ctxKey = "request"

func New[T any](
	log *slog.Logger,
	validate *validator.Validate,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const component = "middleware/request"
			log := log.With(sl.Component(component), sl.ReqID(r))

			var req T
			if err := render.DecodeJSON(r.Body, &req); err != nil {
				var syntaxErr *json.SyntaxError
				var unmarshalTypeErr *json.UnmarshalTypeError

				switch {
				case errors.Is(err, io.EOF):
					log.Warn("request body is empty")
					render.JSON(w, r, response.BadRequest("request body is empty"))

				case errors.Is(err, io.ErrUnexpectedEOF):
					log.Warn("invalid request body")
					render.JSON(w, r, response.BadRequest("invalid request body"))

				case errors.As(err, &syntaxErr):
					log.Warn("invalid json syntax", slog.Int64("offset", syntaxErr.Offset))
					render.JSON(w, r, response.BadRequest("invalid json syntax"))

				case errors.As(err, &unmarshalTypeErr):
					log.Warn("invalid field type",
						slog.String("field", unmarshalTypeErr.Field),
						slog.String("expected", unmarshalTypeErr.Type.String()),
					)
					render.JSON(
						w,
						r,
						response.BadRequest(
							fmt.Sprintf("invalid type for field '%s'", unmarshalTypeErr.Field),
						),
					)

				case uuid.IsInvalidLengthError(err):
					log.Warn("invalid request uuid field length")
					render.JSON(w, r, response.BadRequest("invalid request uuid field length"))

				default:
					log.Error("failed to decode request body", sl.Err(err))
					render.JSON(w, r, response.InternalError())
				}

				return
			}

			log.Debug("request decoded", slog.Any("request", req))

			if err := validate.Struct(req); err != nil {
				if err, ok := errors.AsType[validator.ValidationErrors](err); ok {
					log.Warn("invalid request body", sl.Err(err))
					render.JSON(w, r, response.ValidationError(err))
					return
				}

				log.Error("failed to validate request", sl.Err(err))
				render.JSON(w, r, response.InternalError())
				return
			}

			log.Debug("request validated successfully")

			ctx := context.WithValue(r.Context(), requestKey, req)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext[T any](ctx context.Context) (T, bool) {
	val := ctx.Value(requestKey)
	req, ok := val.(T)
	return req, ok
}

func MustFromContext[T any](ctx context.Context) T {
	const op = "middlewares.request.MustFromContext"

	req, ok := FromContext[T](ctx)
	if !ok {
		panic(
			fmt.Errorf(
				"%s: can not get request from context, looks like you forgot to use middleware",
				op,
			),
		)
	}
	return req
}

func MustGet[T any](r *http.Request) T {
	return MustFromContext[T](r.Context())
}
