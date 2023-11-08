package v1

import (
	"fmt"
	"strings"
	"time"
)

type AuthedRequest struct {
	Username string `header:"X-User-Name" valid:"required"`
}

type PaginatedRequest struct {
	Page uint64 `query:"page" valid:"positive_uint,optional"`
	Size uint64 `query:"size" valid:"range(0|100),optional"`
}

type PaginatedResponse struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"pageSize"`
	Total    uint64 `json:"totalElements"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Library struct {
	ID      string `json:"libraryUid"`
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
}

type Book struct {
	ID        string `json:"bookUid"`
	Name      string `json:"name"`
	Author    string `json:"author"`
	Genre     string `json:"genre"`
	Condition string `json:"condition"`
	Available uint64 `json:"availableCount"`
}

type Rating struct {
	Stars uint64 `json:"stars"`
}

type Time struct {
	time.Time `valid:"required"`
}

func (ct *Time) UnmarshalJSON(b []byte) error {
	var err error

	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}

		return nil
	}

	ct.Time, err = time.Parse(time.DateOnly, s)

	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}

	return nil
}
