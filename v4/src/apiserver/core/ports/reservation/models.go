package reservation

import (
	"time"

	"ds-lab2-bmstu/apiserver/core/ports/library"
	"ds-lab2-bmstu/apiserver/core/ports/rating"
)

type Info struct {
	ID        string
	Username  string
	Status    string
	Start     time.Time
	End       time.Time
	BookID    string
	LibraryID string
}

type FullInfo struct {
	ID           string
	Username     string
	Status       string
	Start        time.Time
	End          time.Time
	ReservedBook library.ReservedBook
	Rating       rating.Rating
}
