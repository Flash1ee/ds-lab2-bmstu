package common

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ds-lab2-bmstu/pkg/readiness"
)

func InitListener(mx *echo.Echo, prober *readiness.Probe) error {
	mx.GET("/liveness", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	mx.GET("/readiness", func(c echo.Context) error {
		if prober.Ready() {
			return c.NoContent(http.StatusOK)
		}

		return c.NoContent(http.StatusServiceUnavailable)
	})

	return nil
}
