package tests

import (
	"database/sql"
	"testing"

	"github.com/makasim/flowstate/sqlitedriver"
	"github.com/makasim/flowstate/usecase"
	"github.com/stretchr/testify/require"
)

func TestDelay_DelayedWin_WithCommit(t *testing.T) {
	db, err := sql.Open("sqlite3", `:memory:`)
	require.NoError(t, err)

	d := sqlitedriver.New(db)

	usecase.Delay_DelayedWin_WithCommit(t, d, d)
}
