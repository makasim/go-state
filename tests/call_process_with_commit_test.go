package tests

import (
	"testing"

	"github.com/makasim/flowstate/memdriver"
	"github.com/makasim/flowstate/usecase"
)

func TestCallProcessWithCommit(t *testing.T) {
	d := memdriver.New()

	testcases.CallProcessWithCommit(t, d, d)
}
