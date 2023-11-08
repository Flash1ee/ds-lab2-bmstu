package httpwrapper

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"ds-lab2-bmstu/pkg/httpvalidator"
)

type ValidationErrorResponse struct {
	Message string                          `json:"message"`
	Errors  []httpvalidator.ValidationError `json:"errors"`
}

func WrapRequest[T any](
	lg *slog.Logger, handler func(echo.Context, T) error,
) func(echo.Context) error {
	return func(c echo.Context) error {
		binder := &echo.DefaultBinder{}

		var req T
		if err := binder.Bind(&req, c); err != nil {
			lg.Warn("failed to bind request", "error", err)

			return c.String(http.StatusBadRequest, "bad request") //nolint: wrapcheck
		}

		if err := binder.BindQueryParams(c, &req); err != nil {
			lg.Warn("failed to bind headers", "error", err)

			return c.String(http.StatusBadRequest, "bad request") //nolint: wrapcheck
		}

		if err := binder.BindHeaders(c, &req); err != nil {
			lg.Warn("failed to bind headers", "error", err)

			return c.String(http.StatusBadRequest, "bad request") //nolint: wrapcheck
		}

		if err := c.Validate(req); err != nil {
			lg.Warn("failed to validate request", "error", err)
			resp := ValidationErrorResponse{
				http.StatusText(http.StatusBadRequest),
				httpvalidator.ParseErrors(err),
			}

			return c.JSON(http.StatusBadRequest, resp)
		}

		return handler(c, req)
	}
}
