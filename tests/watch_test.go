package tests

import (
	"testing"

	"github.com/makasim/flowstate/memdriver"
	"github.com/makasim/flowstate/usecase"
)

func TestWatch(t *testing.T) {
	d := memdriver.New()

	testcases.Watch(t, d, d)
}
