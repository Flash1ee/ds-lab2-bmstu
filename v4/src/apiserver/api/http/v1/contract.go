package v1

import (
	"context"
	"time"

	"ds-lab2-bmstu/apiserver/core/ports/library"
	"ds-lab2-bmstu/apiserver/core/ports/rating"
	"ds-lab2-bmstu/apiserver/core/ports/reservation"
)

type Core interface {
	GetLibraries(context.Context, string, uint64, uint64) (library.Infos, error)
	GetLibraryBooks(context.Context, string, bool, uint64, uint64) (library.Books, error)
	GetUserRating(ctx context.Context, username string) (rating.Rating, error)
	GetUserReservations(context.Context, string) ([]reservation.FullInfo, error)
	TakeBook(ctx context.Context, usename, libraryID, bookID string, end time.Time) (reservation.FullInfo, error)
	ReturnBook(ctx context.Context, username, reservationID, condition string, date time.Time) error
}
