package sqlite_test

import (
	"testing"

	"github.com/pako-23/go-sqlite"
	"github.com/stretchr/testify/require"
)

func TestResultCodeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     sqlite.ResultCode
		expected string
	}{
		{sqlite.ABORT, "ABORT"},
		{sqlite.AUTH, "AUTH"},
		{sqlite.BUSY, "BUSY"},
		{sqlite.CANTOPEN, "CANTOPEN"},
		{sqlite.CONSTRAINT, "CONSTRAINT"},
		{sqlite.CORRUPT, "CORRUPT"},
		{sqlite.DONE, "DONE"},
		{sqlite.EMPTY, "EMPTY"},
		{sqlite.ERROR, "ERROR"},
		{sqlite.FORMAT, "FORMAT"},
		{sqlite.FULL, "FULL"},
		{sqlite.INTERNAL, "INTERNAL"},
		{sqlite.INTERRUPT, "INTERRUPT"},
		{sqlite.IOERR, "IOERR"},
		{sqlite.LOCKED, "LOCKED"},
		{sqlite.MISMATCH, "MISMATCH"},
		{sqlite.MISUSE, "MISUSE"},
		{sqlite.NOLFS, "NOLFS"},
		{sqlite.NOMEM, "NOMEM"},
		{sqlite.NOTADB, "NOTADB"},
		{sqlite.NOTFOUND, "NOTFOUND"},
		{sqlite.NOTICE, "NOTICE"},
		{sqlite.OK, "OK"},
		{sqlite.PERM, "PERM"},
		{sqlite.PROTOCOL, "PROTOCOL"},
		{sqlite.RANGE, "RANGE"},
		{sqlite.READONLY, "READONLY"},
		{sqlite.ROW, "ROW"},
		{sqlite.SCHEMA, "SCHEMA"},
		{sqlite.TOOBIG, "TOOBIG"},
		{sqlite.WARNING, "WARNING"},
		{sqlite.ResultCode(999), "UNKNOWN"},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, test.code.String())
	}
}
