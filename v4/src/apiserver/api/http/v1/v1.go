package v1

import (
	"log/slog"

	"github.com/labstack/echo/v4"

	"ds-lab2-bmstu/pkg/httpwrapper"
)

type api struct {
	core Core
}

func InitListener(mx *echo.Echo, lg *slog.Logger, core Core) error {
	gr := mx.Group("/api/v1")

	a := api{core: core}

	gr.GET("/libraries", httpwrapper.WrapRequest(lg, a.GetLibraries))
	gr.GET("/libraries/:id/books", httpwrapper.WrapRequest(lg, a.GetLibraryBooks))

	gr.GET("/reservations", httpwrapper.WrapRequest(lg, a.GetReservations))
	gr.POST("/reservations", httpwrapper.WrapRequest(lg, a.TakeBook))
	gr.POST("/reservations/:id/return", httpwrapper.WrapRequest(lg, a.ReturnBook))

	gr.GET("/rating", httpwrapper.WrapRequest(lg, a.GetRating))

	return nil
}
