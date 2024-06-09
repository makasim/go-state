package tests

import (
	"testing"

	"github.com/makasim/flowstate/memdriver"
	"github.com/makasim/flowstate/usecase"
)

func TestForkJoin_FirstWins(t *testing.T) {
	d := memdriver.New()

	usecase.ForkJoin_FirstWins(t, d, d)
}
