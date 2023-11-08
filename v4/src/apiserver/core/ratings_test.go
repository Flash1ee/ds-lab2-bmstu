//go:build testing
// +build testing

package core

import (
	"context"
	"testing"

	"ds-lab2-bmstu/apiserver/core/ports/rating"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RatingSuite struct {
	TestSuite
}

func (s *RatingSuite) SetupTest() {
	s.TestSuite.SetupTest()
}

func (s *RatingSuite) TearDownTest() {
	s.TestSuite.TearDownTest()
}

func (s *RatingSuite) TestGet() {
	cases := map[string]struct {
		username string
		res      rating.Rating
		err      error
	}{
		"normal case": {
			username: "qwerty",
			res: rating.Rating{
				Stars: 10,
			},
		},
	}

	for tn, tc := range cases {
		s.T().Run(tn, func(t *testing.T) {
			ctx := context.Background()

			s.mockedRating.EXPECT().GetUserRating(ctx, tc.username).Return(tc.res, tc.err)

			res, err := s.core.GetUserRating(ctx, tc.username)
			require.Equal(s.T(), tc.res, res)
			require.ErrorIs(s.T(), err, tc.err)
		})
	}
}

func TestRatingSuite(t *testing.T) {
	suite.Run(t, new(RatingSuite))
}
