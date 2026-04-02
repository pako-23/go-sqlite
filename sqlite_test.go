package sqlite_test

import (
	"testing"

	"github.com/pako-23/go-sqlite"
	"github.com/stretchr/testify/require"
)

func TestResultCodeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     int
		expected string
	}{
		{sqlite.ResultCodeAbort, "ABORT"},
		{sqlite.ResultCodeAuth, "AUTH"},
		{sqlite.ResultCodeBusy, "BUSY"},
		{sqlite.ResultCodeCantOpen, "CANTOPEN"},
		{sqlite.ResultCodeConstraint, "CONSTRAINT"},
		{sqlite.ResultCodeCorrupt, "CORRUPT"},
		{sqlite.ResultCodeDone, "DONE"},
		{sqlite.ResultCodeEmpty, "EMPTY"},
		{sqlite.ResultCodeError, "ERROR"},
		{sqlite.ResultCodeFormat, "FORMAT"},
		{sqlite.ResultCodeFull, "FULL"},
		{sqlite.ResultCodeInternal, "INTERNAL"},
		{sqlite.ResultCodeInterrupt, "INTERRUPT"},
		{sqlite.ResultCodeIOErr, "IOERR"},
		{sqlite.ResultCodeLocked, "LOCKED"},
		{sqlite.ResultCodeMismatch, "MISMATCH"},
		{sqlite.ResultCodeMisuse, "MISUSE"},
		{sqlite.ResultCodeNoLFS, "NOLFS"},
		{sqlite.ResultCodeNoMem, "NOMEM"},
		{sqlite.ResultCodeNotADB, "NOTADB"},
		{sqlite.ResultCodeNotFound, "NOTFOUND"},
		{sqlite.ResultCodeNotice, "NOTICE"},
		{sqlite.ResultCodeOK, "OK"},
		{sqlite.ResultCodePerm, "PERM"},
		{sqlite.ResultCodeProtocol, "PROTOCOL"},
		{sqlite.ResultCodeRange, "RANGE"},
		{sqlite.ResultCodeReadOnly, "READONLY"},
		{sqlite.ResultCodeRow, "ROW"},
		{sqlite.ResultCodeSchema, "SCHEMA"},
		{sqlite.ResultCodeTooBig, "TOOBIG"},
		{sqlite.ResultCodeWarning, "WARNING"},
		{999, "UNKNOWN"},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, sqlite.ResultCodeText(test.code))
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	err := &sqlite.Error{
		Code:    sqlite.ResultCodeCantOpen,
		Message: "unable to open database file",
	}

	require.Equal(t, "sqlite3 error CANTOPEN(14): unable to open database file", err.Error())

}

func TestOpenCloseSuccess(t *testing.T) {
	t.Parallel()

	conn, err := sqlite.Open(":memory:")

	require.NoError(t, err)
	require.NotNil(t, conn)

	err = conn.Close()
	require.NoError(t, err)
}

func TestOpenFailure(t *testing.T) {
	t.Parallel()

	_, err := sqlite.Open("file:/this/path/does/not/exist/db.sqlite")

	require.Error(t, err)
	require.EqualError(t, err, "sqlite3 error CANTOPEN(14): unable to open database file")
}

func TestSimpleQuery(t *testing.T) {
	t.Parallel()

	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	query := "SELECT 100 AS id, 'name' AS name, 3.14 AS pi, x'0102' AS data, NULL AS empty"
	statement, err := conn.Prepare(query)
	require.NoError(t, err)
	require.NotNil(t, statement)

	require.Equal(t, 5, statement.ColumnCount())

	done, err := statement.Step()
	require.NoError(t, err)
	require.False(t, done)

	value, err := statement.Column(0)
	require.NoError(t, err)
	require.Equal(t, int64(100), value)

	value, err = statement.Column(1)
	require.NoError(t, err)
	require.Equal(t, "name", value)

	value, err = statement.Column(2)
	require.NoError(t, err)
	require.InDelta(t, 3.14, value, 0.0001)

	value, err = statement.Column(3)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0x02}, value)

	value, err = statement.Column(4)
	require.NoError(t, err)
	require.Nil(t, value)

	done, err = statement.Step()
	require.NoError(t, err)
	require.True(t, done)

	require.NoError(t, statement.Finalize())
}

func TestSimpleQueryRow(t *testing.T) {
	t.Parallel()

	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	query := "SELECT 100 AS id, 'name' AS name, 3.14 AS pi, x'0102' AS data, NULL AS empty"
	statement, err := conn.Prepare(query)
	require.NoError(t, err)
	require.NotNil(t, statement)

	require.Equal(t, 5, statement.ColumnCount())

	done, err := statement.Step()
	require.NoError(t, err)
	require.False(t, done)

	row, err := statement.Row()
	require.NoError(t, err)

	require.Len(t, row, 5)

	require.Equal(t, int64(100), row[0])
	require.Equal(t, "name", row[1])
	require.InDelta(t, 3.14, row[2], 0.0001)
	require.Equal(t, []byte{0x01, 0x02}, row[3])
	require.Nil(t, row[4])

	done, err = statement.Step()
	require.NoError(t, err)
	require.True(t, done)

	require.NoError(t, statement.Finalize())
}

func TestStatementErrors(t *testing.T) {
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	t.Run("not existing table", func(t *testing.T) {
		statement, err := conn.Prepare("SELECT * FROM non_existent_table")
		require.Error(t, err)
		require.Nil(t, statement)
		require.EqualError(t, err, "sqlite3 error ERROR(1): no such table: non_existent_table")
	})

	t.Run("column out of bounds", func(t *testing.T) {
		statement, err := conn.Prepare("SELECT 1")
		require.NoError(t, err)
		defer statement.Finalize()

		_, err = statement.Step()
		require.NoError(t, err)

		_, err = statement.Column(1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid column number")
	})
}
