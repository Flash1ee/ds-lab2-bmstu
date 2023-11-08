package v1

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"ds-lab2-bmstu/apiserver/core"
)

type ReturnBookRequest struct {
	AuthedRequest `valid:"optional"`
	ID            string `param:"id" valid:"uuidv4,required"`
	Condition     string `json:"condition" valid:"optional"`
	Date          Time   `json:"date" valid:"optional"`
}

func (a *api) ReturnBook(c echo.Context, req ReturnBookRequest) error {
	err := a.core.ReturnBook(c.Request().Context(), req.Username, req.ID, req.Condition, req.Date.Time)
	if err == nil {
		return c.NoContent(http.StatusNoContent)

	}
	if errors.Is(err, core.ErrNotFound) {
		resp := ErrorResponse{
			Message: "no such reservation",
		}
		return c.JSON(http.StatusNotFound, &resp)
	}
	return c.NoContent(http.StatusInternalServerError)
}
