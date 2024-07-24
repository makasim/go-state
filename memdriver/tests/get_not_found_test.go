package tests

import (
	"testing"

	"github.com/makasim/flowstate/memdriver"
	"github.com/makasim/flowstate/testcases"
)

func TestGetNotFound(t *testing.T) {
	d := memdriver.New()

	testcases.GetNotFound(t, d, d)
}
