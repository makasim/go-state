package tests

import (
	"testing"

	"github.com/makasim/flowstate"
	"github.com/makasim/flowstate/memdriver"
	"github.com/stretchr/testify/require"
)

func TestThreeConsequentNodes(t *testing.T) {
	trkr := &tracker2{}

	br := &flowstate.MapBehaviorRegistry{}
	br.SetBehavior("first", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track2(taskCtx, trkr)
		return flowstate.Transit(taskCtx, `second`), nil
	}))
	br.SetBehavior("second", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track2(taskCtx, trkr)
		return flowstate.Transit(taskCtx, `third`), nil
	}))
	br.SetBehavior("third", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track2(taskCtx, trkr)
		return flowstate.End(taskCtx), nil
	}))

	d := &memdriver.Driver{}
	e := flowstate.NewEngine(d, br)

	taskCtx := &flowstate.TaskCtx{
		Current: flowstate.Task{
			ID:  "aTID",
			Rev: 0,
		},
	}

	require.NoError(t, e.Do(flowstate.Transit(taskCtx, `first`)))
	require.NoError(t, e.Execute(taskCtx))

	require.Equal(t, []string{`first`, `second`, `third`}, trkr.Visited())
}
