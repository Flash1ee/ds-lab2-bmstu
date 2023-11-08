//go:build testing
// +build testing

package core

import (
	"io"
	"log/slog"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"ds-lab2-bmstu/apiserver/core/ports/library"
	"ds-lab2-bmstu/apiserver/core/ports/rating"
	"ds-lab2-bmstu/apiserver/core/ports/reservation"
	"ds-lab2-bmstu/pkg/readiness"
)

type TestSuite struct {
	suite.Suite

	core *Core

	mockedLibrary     *library.MockClient
	mockedRating      *rating.MockClient
	mockedReservation *reservation.MockClient
}

func (s *TestSuite) SetupTest() {
	s.mockedLibrary = library.NewMockClient(s.T())
	s.mockedReservation = reservation.NewMockClient(s.T())
	s.mockedRating = rating.NewMockClient(s.T())

	var err error
	s.core, err = New(
		slog.New(slog.NewJSONHandler(io.Discard, nil)), readiness.New(),
		s.mockedLibrary, s.mockedRating, s.mockedReservation,
	)

	require.NoError(s.T(), err, "failed to init core")
}

func (s *TestSuite) TearDownTest() {
}
