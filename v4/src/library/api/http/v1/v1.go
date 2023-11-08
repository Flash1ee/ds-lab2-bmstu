package v1

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"

	"ds-lab2-bmstu/library/core/ports/libraries"
	"ds-lab2-bmstu/pkg/httpwrapper"
)

type Core interface {
	GetLibraryBooks(context.Context, string, bool, uint64, uint64) (libraries.LibraryBooks, error)
	GetLibraryBooksByIDs(context.Context, []string) (libraries.LibraryBooks, error)
	GetLibraries(context.Context, string, uint64, uint64) (libraries.Libraries, error)
	GetLibrariesByIDs(context.Context, []string) (libraries.Libraries, error)
	TakeBook(context.Context, string, string) (libraries.ReservedBook, error)
	ReturnBook(context.Context, string, string) (libraries.Book, error)
}

type api struct {
	core Core
}

func InitListener(mx *echo.Echo, lg *slog.Logger, core Core) error {
	gr := mx.Group("/api/v1")

	a := api{core: core}

	gr.GET("/libraries", httpwrapper.WrapRequest(lg, a.GetLibraries))
	gr.GET("/libraries/:id/books", httpwrapper.WrapRequest(lg, a.GetLibraryBooks))
	gr.POST("/libraries/:lib_id/books/:book_id/return", httpwrapper.WrapRequest(lg, a.ReturnBook))

	gr.POST("/books", httpwrapper.WrapRequest(lg, a.TakeBook))
	gr.GET("/books", httpwrapper.WrapRequest(lg, a.GetLibraryBooks))

	return nil
}
