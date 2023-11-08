package v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type ReservationsRequest struct {
	AuthedRequest `valid:"optional"`

	Status string `query:"status" valid:"in(RENTED|RETURNED|EXPIRED),optional"`
}

type Reservation struct {
	ID        string    `json:"reservationUid" valid:"uuidv4,required"`
	Status    string    `json:"status" valid:"in(RENTED|RETURNED|EXPIRED)"`
	Start     time.Time `json:"startDate"`
	End       time.Time `json:"tillDate"`
	BookID    string    `json:"bookId"`
	LibraryID string    `json:"libraryId"`
}

func (a *api) GetReservations(c echo.Context, req ReservationsRequest) error {
	data, err := a.core.GetUserReservations(c.Request().Context(), req.Username, req.Status)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	var resp []Reservation
	for _, res := range data {
		resp = append(resp, Reservation(res))
	}

	return c.JSON(http.StatusOK, &resp)
}
