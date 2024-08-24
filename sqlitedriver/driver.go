package sqlitedriver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/makasim/flowstate"
	"github.com/makasim/flowstate/memdriver"
	"github.com/makasim/flowstate/stddoer"

	_ "github.com/mattn/go-sqlite3"
)

type Driver struct {
	*memdriver.FlowRegistry
	db    *sql.DB
	doers []flowstate.Doer
}

func New(db *sql.DB) *Driver {
	d := &Driver{
		FlowRegistry: &memdriver.FlowRegistry{},

		db: db,
	}

	dlr := NewDelayer(d.db)
	dl := NewDataLog(d.db)

	d.doers = []flowstate.Doer{
		stddoer.Transit(),
		stddoer.Pause(),
		stddoer.Resume(),
		stddoer.End(),
		stddoer.Noop(),
		stddoer.Recoverer(time.Millisecond * 500),
		stddoer.NewSerializer(),
		stddoer.NewDeserializer(),
		flowstate.DefaultReferenceDataDoer,
		flowstate.DefaultDereferenceDataDoer,

		memdriver.NewFlowGetter(d.FlowRegistry),

		NewCommiter(d.db, dlr, dl),
		NewWatcher(d.db),
		NewGetter(d.db),
		dlr,
		dl,
	}

	return d
}

func (d *Driver) Do(cmd0 flowstate.Command) error {
	for _, doer := range d.doers {
		if err := doer.Do(cmd0); errors.Is(err, flowstate.ErrCommandNotSupported) {
			continue
		} else if err != nil {
			return fmt.Errorf("%T: do: %w", doer, err)
		}

		return nil
	}

	return fmt.Errorf("no doer for command %T", cmd0)
}

func (d *Driver) Init(e *flowstate.Engine) error {
	if _, err := d.db.Exec(createRevTableSQL); err != nil {
		return fmt.Errorf("create flowstate_rev table: db: exec: %w", err)
	}
	if _, err := d.db.Exec(createStateLatestTableSQL); err != nil {
		return fmt.Errorf("create flowstate_state_latest table: db: exec: %w", err)
	}
	if _, err := d.db.Exec(createStateLogTableSQL); err != nil {
		return fmt.Errorf("create flowstate_state_log table: db: exec: %w", err)
	}
	if _, err := d.db.Exec(createDataLogTableSQL); err != nil {
		return fmt.Errorf("create flowstate_data_log table: db: exec: %w", err)
	}

	for _, doer := range d.doers {
		if err := doer.Init(e); err != nil {
			return fmt.Errorf("%T: init: %w", doer, err)
		}
	}
	return nil
}

func (d *Driver) Shutdown(ctx context.Context) error {
	var res error
	for _, doer := range d.doers {
		if err := doer.Shutdown(ctx); err != nil {
			res = errors.Join(res, fmt.Errorf("%T: shutdown: %w", doer, err))
		}
	}

	return res
}
