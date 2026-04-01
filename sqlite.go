package sqlite

/*
#include <stdlib.h>
#include "sqlite3.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	ResultCodeAbort      = C.SQLITE_ABORT
	ResultCodeAuth       = C.SQLITE_AUTH
	ResultCodeBusy       = C.SQLITE_BUSY
	ResultCodeCantOpen   = C.SQLITE_CANTOPEN
	ResultCodeConstraint = C.SQLITE_CONSTRAINT
	ResultCodeCorrupt    = C.SQLITE_CORRUPT
	ResultCodeDone       = C.SQLITE_DONE
	ResultCodeEmpty      = C.SQLITE_EMPTY
	ResultCodeError      = C.SQLITE_ERROR
	ResultCodeFormat     = C.SQLITE_FORMAT
	ResultCodeFull       = C.SQLITE_FULL
	ResultCodeInternal   = C.SQLITE_INTERNAL
	ResultCodeInterrupt  = C.SQLITE_INTERRUPT
	ResultCodeIOErr      = C.SQLITE_IOERR
	ResultCodeLocked     = C.SQLITE_LOCKED
	ResultCodeMismatch   = C.SQLITE_MISMATCH
	ResultCodeMisuse     = C.SQLITE_MISUSE
	ResultCodeNoLFS      = C.SQLITE_NOLFS
	ResultCodeNoMem      = C.SQLITE_NOMEM
	ResultCodeNotADB     = C.SQLITE_NOTADB
	ResultCodeNotFound   = C.SQLITE_NOTFOUND
	ResultCodeNotice     = C.SQLITE_NOTICE
	ResultCodeOK         = C.SQLITE_OK
	ResultCodePerm       = C.SQLITE_PERM
	ResultCodeProtocol   = C.SQLITE_PROTOCOL
	ResultCodeRange      = C.SQLITE_RANGE
	ResultCodeReadOnly   = C.SQLITE_READONLY
	ResultCodeRow        = C.SQLITE_ROW
	ResultCodeSchema     = C.SQLITE_SCHEMA
	ResultCodeTooBig     = C.SQLITE_TOOBIG
	ResultCodeWarning    = C.SQLITE_WARNING
)

func ResultCodeText(code int) string {
	switch code {
	case ResultCodeAbort:
		return "ABORT"
	case ResultCodeAuth:
		return "AUTH"
	case ResultCodeBusy:
		return "BUSY"
	case ResultCodeCantOpen:
		return "CANTOPEN"
	case ResultCodeConstraint:
		return "CONSTRAINT"
	case ResultCodeCorrupt:
		return "CORRUPT"
	case ResultCodeDone:
		return "DONE"
	case ResultCodeEmpty:
		return "EMPTY"
	case ResultCodeError:
		return "ERROR"
	case ResultCodeFormat:
		return "FORMAT"
	case ResultCodeFull:
		return "FULL"
	case ResultCodeInternal:
		return "INTERNAL"
	case ResultCodeInterrupt:
		return "INTERRUPT"
	case ResultCodeIOErr:
		return "IOERR"
	case ResultCodeLocked:
		return "LOCKED"
	case ResultCodeMismatch:
		return "MISMATCH"
	case ResultCodeMisuse:
		return "MISUSE"
	case ResultCodeNoLFS:
		return "NOLFS"
	case ResultCodeNoMem:
		return "NOMEM"
	case ResultCodeNotADB:
		return "NOTADB"
	case ResultCodeNotFound:
		return "NOTFOUND"
	case ResultCodeNotice:
		return "NOTICE"
	case ResultCodeOK:
		return "OK"
	case ResultCodePerm:
		return "PERM"
	case ResultCodeProtocol:
		return "PROTOCOL"
	case ResultCodeRange:
		return "RANGE"
	case ResultCodeReadOnly:
		return "READONLY"
	case ResultCodeRow:
		return "ROW"
	case ResultCodeSchema:
		return "SCHEMA"
	case ResultCodeTooBig:
		return "TOOBIG"
	case ResultCodeWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

const (
	DataTypeInteger = C.SQLITE_INTEGER
	DataTypeFloat   = C.SQLITE_FLOAT
	DataTypeBlob    = C.SQLITE_BLOB
	DataTypeNull    = C.SQLITE_NULL
	DataTypeText    = C.SQLITE3_TEXT
)

type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	return fmt.Sprintf("sqlite3 error %s(%d): %s", ResultCodeText(e.Code), e.Code, e.Message)
}

type Conn struct {
	conn *C.sqlite3
}

type Statement struct {
	columnCount int
	statement   *C.sqlite3_stmt
}

func Open(filename string) (*Conn, error) {
	filenamePtr := C.CString(filename)
	defer C.free(unsafe.Pointer(filenamePtr))

	var conn *C.sqlite3

	rv := C.sqlite3_open(filenamePtr, &conn)
	if rv != C.SQLITE_OK {
		err := &Error{
			Code:    int(rv),
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
			Code:    int(rv),
			Message: C.GoString(C.sqlite3_errmsg(c.conn)),
		}
	}

	return nil
}

func (c *Conn) Prepare(query string) (*Statement, error) {
	queryPtr := C.CString(query)
	defer C.free(unsafe.Pointer(queryPtr))

	var statement *C.sqlite3_stmt

	rv := C.sqlite3_prepare_v2(c.conn, queryPtr, C.int(len(query)+1), &statement, nil)
	if rv != C.SQLITE_OK {
		return nil, &Error{
			Code:    int(rv),
			Message: C.GoString(C.sqlite3_errstr(rv)),
		}
	}

	return &Statement{
		columnCount: int(C.sqlite3_column_count(statement)),
		statement:   statement,
	}, nil
}

func (s *Statement) ColumnCount() int {
	return s.columnCount
}

func (s *Statement) Finalize() error {
	rv := C.sqlite3_finalize(s.statement)
	if rv != C.SQLITE_OK {
		return &Error{
			Code:    int(rv),
			Message: C.GoString(C.sqlite3_errstr(rv)),
		}
	}

	return nil
}

func (s *Statement) Step() (bool, error) {
	rv := C.sqlite3_step(s.statement)
	switch rv {
	case C.SQLITE_ROW:
		return false, nil

	case C.SQLITE_DONE:
		return true, nil

	default:
		return true, &Error{
			Code:    int(rv),
			Message: C.GoString(C.sqlite3_errstr(rv)),
		}
	}
}

func (s *Statement) getColumn(i int) any {
	columnType := int(C.sqlite3_column_type(s.statement, C.int(i)))
	switch columnType {
	case DataTypeInteger:
		return int64(C.sqlite3_column_int64(s.statement, C.int(i)))

	case DataTypeFloat:
		return float64(C.sqlite3_column_double(s.statement, C.int(i)))

	case DataTypeBlob:
		ptr := C.sqlite3_column_blob(s.statement, C.int(i))
		size := C.sqlite3_column_bytes(s.statement, C.int(i))
		return C.GoBytes(ptr, size)

	case DataTypeText:
		ptr := C.sqlite3_column_text(s.statement, C.int(i))
		return C.GoString((*C.char)(unsafe.Pointer(ptr)))

	default:
		return nil

	}
}

func (s *Statement) Column(i int) (any, error) {
	if i < 0 || i >= s.ColumnCount() {
		return nil, errors.New("invalid column number")
	}

	return s.getColumn(i), nil
}

func (s *Statement) Row() ([]any, error) {
	row := make([]any, s.ColumnCount())

	for i := range s.ColumnCount() {
		row[i] = s.getColumn(i)
	}

	return row, nil
}
