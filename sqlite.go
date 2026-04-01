package sqlite

/*
#include <stdlib.h>
#include "sqlite3.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type ResultCode int

const (
	ABORT      ResultCode = C.SQLITE_ABORT
	AUTH       ResultCode = C.SQLITE_AUTH
	BUSY       ResultCode = C.SQLITE_BUSY
	CANTOPEN   ResultCode = C.SQLITE_CANTOPEN
	CONSTRAINT ResultCode = C.SQLITE_CONSTRAINT
	CORRUPT    ResultCode = C.SQLITE_CORRUPT
	DONE       ResultCode = C.SQLITE_DONE
	EMPTY      ResultCode = C.SQLITE_EMPTY
	ERROR      ResultCode = C.SQLITE_ERROR
	FORMAT     ResultCode = C.SQLITE_FORMAT
	FULL       ResultCode = C.SQLITE_FULL
	INTERNAL   ResultCode = C.SQLITE_INTERNAL
	INTERRUPT  ResultCode = C.SQLITE_INTERRUPT
	IOERR      ResultCode = C.SQLITE_IOERR
	LOCKED     ResultCode = C.SQLITE_LOCKED
	MISMATCH   ResultCode = C.SQLITE_MISMATCH
	MISUSE     ResultCode = C.SQLITE_MISUSE
	NOLFS      ResultCode = C.SQLITE_NOLFS
	NOMEM      ResultCode = C.SQLITE_NOMEM
	NOTADB     ResultCode = C.SQLITE_NOTADB
	NOTFOUND   ResultCode = C.SQLITE_NOTFOUND
	NOTICE     ResultCode = C.SQLITE_NOTICE
	OK         ResultCode = C.SQLITE_OK
	PERM       ResultCode = C.SQLITE_PERM
	PROTOCOL   ResultCode = C.SQLITE_PROTOCOL
	RANGE      ResultCode = C.SQLITE_RANGE
	READONLY   ResultCode = C.SQLITE_READONLY
	ROW        ResultCode = C.SQLITE_ROW
	SCHEMA     ResultCode = C.SQLITE_SCHEMA
	TOOBIG     ResultCode = C.SQLITE_TOOBIG
	WARNING    ResultCode = C.SQLITE_WARNING
)

func (r ResultCode) String() string {
	names := map[ResultCode]string{
		ABORT:      "ABORT",
		AUTH:       "AUTH",
		BUSY:       "BUSY",
		CANTOPEN:   "CANTOPEN",
		CONSTRAINT: "CONSTRAINT",
		CORRUPT:    "CORRUPT",
		DONE:       "DONE",
		EMPTY:      "EMPTY",
		ERROR:      "ERROR",
		FORMAT:     "FORMAT",
		FULL:       "FULL",
		INTERNAL:   "INTERNAL",
		INTERRUPT:  "INTERRUPT",
		IOERR:      "IOERR",
		LOCKED:     "LOCKED",
		MISMATCH:   "MISMATCH",
		MISUSE:     "MISUSE",
		NOLFS:      "NOLFS",
		NOMEM:      "NOMEM",
		NOTADB:     "NOTADB",
		NOTFOUND:   "NOTFOUND",
		NOTICE:     "NOTICE",
		OK:         "OK",
		PERM:       "PERM",
		PROTOCOL:   "PROTOCOL",
		RANGE:      "RANGE",
		READONLY:   "READONLY",
		ROW:        "ROW",
		SCHEMA:     "SCHEMA",
		TOOBIG:     "TOOBIG",
		WARNING:    "WARNING",
	}

	name, ok := names[r]
	if !ok {
		return "UNKNOWN"
	}

	return name
}

type Error struct {
	Message string
	Code    ResultCode
}

func (e *Error) Error() string {
	return fmt.Sprintf("sqlite3 error %s(%d): %s", e.Code.String(), e.Code, e.Message)
}

type Conn struct {
	conn *C.sqlite3
}

func Open(filename string) (*Conn, error) {
	filenamePtr := C.CString(filename)
	defer C.free(unsafe.Pointer(filenamePtr))

	var conn *C.sqlite3

	rv := C.sqlite3_open(filenamePtr, &conn)
	if rv != C.SQLITE_OK {
		err := &Error{
			Code:    ResultCode(rv),
			Message: C.GoString(C.sqlite3_errmsg(conn)),
		}

		C.sqlite3_close(conn)
		return nil, err
	}

	return &Conn{conn: conn}, nil
}

func (c *Conn) Close() error {
	rv := C.sqlite3_close(c.conn)
	if rv != C.SQLITE_OK {
		return &Error{
			Code:    ResultCode(rv),
			Message: C.GoString(C.sqlite3_errmsg(c.conn)),
		}
	}

	return nil
}
