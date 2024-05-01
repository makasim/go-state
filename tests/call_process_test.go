package tests

import (
	"testing"
	"time"

	"github.com/makasim/flowstate"
	"github.com/makasim/flowstate/memdriver"
	"github.com/stretchr/testify/require"
)

func TestCallProcess(t *testing.T) {
	p := flowstate.Process{
		ID:  "simplePID",
		Rev: 1,
		Nodes: []flowstate.Node{
			{
				ID:         "callNID",
				BehaviorID: "call",
			},
			{
				ID:         "calledNID",
				BehaviorID: "called",
			},
			{
				ID:         "endNID",
				BehaviorID: "end",
			},
		},
		Transitions: []flowstate.Transition{
			{
				ID:     "callTID",
				FromID: "",
				ToID:   "callNID",
			},
			{
				ID:     "calledTID",
				FromID: "",
				ToID:   "calledNID",
			},
			{
				ID:     "callEndTID",
				FromID: "callNID",
				ToID:   "endNID",
			},
			{
				ID:     "calledEndTID",
				FromID: "calledNID",
				ToID:   "endNID",
			},
		},
	}

	var nextTaskCtx *flowstate.TaskCtx
	taskCtx := &flowstate.TaskCtx{
		Current: flowstate.Task{
			ID:         "aTID",
			Rev:        0,
			ProcessID:  p.ID,
			ProcessRev: p.Rev,

			Transition: p.Transitions[0],
		},
		Process: p,
		Node:    p.Nodes[0],
	}

	trkr := &tracker{}

	br := &flowstate.MapBehaviorRegistry{}
	br.SetBehavior("call", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track(taskCtx, trkr)

		if flowstate.Resumed(taskCtx) {
			return flowstate.Transit(taskCtx, `callEndTID`), nil
		}

		nextTaskCtx = &flowstate.TaskCtx{
			Current: flowstate.Task{
				ID:         "aTID",
				Rev:        0,
				ProcessID:  p.ID,
				ProcessRev: p.Rev,

				Transition: p.Transitions[1],
			},
			Process: p,
			Node:    p.Nodes[1],
		}

		if err := taskCtx.Engine.Do(
			flowstate.Pause(taskCtx),
			flowstate.Stack(taskCtx, nextTaskCtx),
			flowstate.Transit(nextTaskCtx, `calledTID`),
		); err != nil {
			return nil, err
		}

		return flowstate.Nop(taskCtx), nil
	}))
	br.SetBehavior("called", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track(taskCtx, trkr)
		return flowstate.Transit(taskCtx, `calledEndTID`), nil
	}))
	br.SetBehavior("end", flowstate.BehaviorFunc(func(taskCtx *flowstate.TaskCtx) (flowstate.Command, error) {
		track(taskCtx, trkr)

		if flowstate.Stacked(taskCtx) {
			callTaskCtx := &flowstate.TaskCtx{}

			if err := taskCtx.Engine.Do(
				flowstate.Unstack(taskCtx, callTaskCtx),
				flowstate.Resume(callTaskCtx),
				flowstate.End(taskCtx),
			); err != nil {
				return nil, err
			}

			return flowstate.Nop(taskCtx), nil
		}

		return flowstate.End(taskCtx), nil
	}))

	d := &memdriver.Driver{}
	e := flowstate.NewEngine(d, br)

	err := e.Execute(taskCtx)
	require.NoError(t, err)

	time.Sleep(time.Second * 5)

	require.Equal(t, []flowstate.TransitionID{
		`callTID`,
		`calledTID`,
		`calledEndTID`,
		`callTID`,
		`callEndTID`,
	}, trkr.Visited())
}
