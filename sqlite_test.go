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
