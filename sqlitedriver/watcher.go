package sqlitedriver

import (
	"context"
	"database/sql"

	"github.com/makasim/flowstate"
)

var _ flowstate.Doer = &Watcher{}

type Watcher struct {
	db *sql.DB
	e  *flowstate.Engine
}

func NewWatcher(db *sql.DB) *Watcher {
	d := &Watcher{
		db: db,
	}

	return d
}

func (w *Watcher) Do(cmd0 flowstate.Command) error {
	cmd, ok := cmd0.(*flowstate.GetWatcherCommand)
	if !ok {
		return flowstate.ErrCommandNotSupported
	}

	lis := &listener{
		db: w.db,

		sinceRev:    cmd.SinceRev,
		sinceLatest: cmd.SinceLatest,
		// todo: copy labels
		labels: cmd.Labels,

		watchCh:  make(chan *flowstate.StateCtx, 1),
		changeCh: make(chan int64, 1),
		closeCh:  make(chan struct{}),
	}

	lis.Change(cmd.SinceRev)

	go lis.listen()

	cmd.Watcher = lis

	return nil
}

func (w *Watcher) Init(e *flowstate.Engine) error {
	w.e = e
	return nil
}

func (w *Watcher) Shutdown(_ context.Context) error {
	return nil
}

type listener struct {
	db *sql.DB

	sinceRev    int64
	sinceLatest bool

	labels   map[string]string
	watchCh  chan *flowstate.StateCtx
	changeCh chan int64

	closeCh chan struct{}
}

func (lis *listener) Watch() <-chan *flowstate.StateCtx {
	return lis.watchCh
}

func (lis *listener) Close() {
	close(lis.closeCh)
}

func (lis *listener) Change(rev int64) {
	select {
	case lis.changeCh <- rev:
	case <-lis.changeCh:
		lis.changeCh <- rev
	}
}

func (lis *listener) listen() {
	var states []*flowstate.StateCtx

	if lis.sinceLatest {
		lis.l.Lock()
		_, sinceRev := lis.l.LatestByLabels(lis.labels)
		lis.sinceRev = sinceRev - 1
		lis.l.Unlock()
	}

skip:
	for {
		select {
		case <-lis.changeCh:
			lis.l.Lock()
			states, lis.sinceRev = lis.l.Entries(lis.sinceRev, 10)
			lis.l.Unlock()

			if len(states) == 0 {
				continue skip
			}

		next:
			for _, t := range states {
				for k, v := range lis.labels {
					if t.Committed.Labels[k] != v {
						continue next
					}
				}

				select {
				case lis.watchCh <- t:
					continue next
				case <-lis.closeCh:
					return
				}
			}
		case <-lis.closeCh:
			lis.l.UnsubscribeCommit(lis.changeCh)
			return
		}
	}
}
