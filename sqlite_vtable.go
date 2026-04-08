package sqlite

/*
#include "gosqlite_vtable.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

type VirtualTableCursor interface {
	Close() error
	Filter(indexId int, indexName string, values []any) error
	Next() error
	EOF() bool
	Column(column int) error
	Rowid() (int64, error)
}

const (
	IndexConstraintEq        = C.SQLITE_INDEX_CONSTRAINT_EQ
	IndexConstraintGt        = C.SQLITE_INDEX_CONSTRAINT_GT
	IndexConstraintLe        = C.SQLITE_INDEX_CONSTRAINT_LE
	IndexConstraintLt        = C.SQLITE_INDEX_CONSTRAINT_LT
	IndexConstraintGe        = C.SQLITE_INDEX_CONSTRAINT_GE
	IndexConstraintMatch     = C.SQLITE_INDEX_CONSTRAINT_MATCH
	IndexConstraintLike      = C.SQLITE_INDEX_CONSTRAINT_LIKE
	IndexConstraintGlob      = C.SQLITE_INDEX_CONSTRAINT_GLOB
	IndexConstraintRegExp    = C.SQLITE_INDEX_CONSTRAINT_REGEXP
	IndexConstraintNe        = C.SQLITE_INDEX_CONSTRAINT_NE
	IndexConstraintIsNot     = C.SQLITE_INDEX_CONSTRAINT_ISNOT
	IndexConstraintIsNotNull = C.SQLITE_INDEX_CONSTRAINT_ISNOTNULL
	IndexConstraintIsNull    = C.SQLITE_INDEX_CONSTRAINT_ISNULL
	IndexConstraintIs        = C.SQLITE_INDEX_CONSTRAINT_IS
	IndexConstraintLimit     = C.SQLITE_INDEX_CONSTRAINT_LIMIT
	IndexConstraintOffset    = C.SQLITE_INDEX_CONSTRAINT_OFFSET
	IndexConstraintFunction  = C.SQLITE_INDEX_CONSTRAINT_FUNCTION
	IndexScanUnique          = C.SQLITE_INDEX_SCAN_UNIQUE
)

const (
	OrderByAsc  uint8 = 0
	OrderByDesc       = 1
)

type IndexConstraint struct {
	Column   int
	Operator uint8
	Usable   bool
}

type IndexOrderBy struct {
	Column    int
	Direction uint8
}

type IndexResult struct {
	Usage           []bool
	IndexNumber     int
	IndexName       string
	OrderByConsumed bool
	EstimatedCost   float64
	EstimatedRows   int64
}

type EponymousVirtualTable interface {
	BestIndex(constraints []IndexConstraint, order []IndexOrderBy) (IndexResult, error)
	Disconnect() error
	Open() (VirtualTableCursor, error)
}

type VirtualTable interface {
	EponymousVirtualTable
	Destroy() error
}

type EponymousModule interface {
	Connect() (EponymousVirtualTable, error)
	Declaration() string
}

type Module interface {
	EponymousModule
	Create() (VirtualTable, error)
}

func sqliteString(s string) *C.char {
	ret := C.sqlite3_malloc(C.int(len(s) + 1))
	if ret == nil {
		panic("failed to allocate sqlite3 string: out of memory")
	}

	slice := unsafe.Slice((*byte)(ret), len(s)+1)
	copy(slice, s)
	slice[len(s)-1] = 0

	return (*C.char)(ret)
}

func gosqliteError(err error, errmsg **C.char) C.int {
	var (
		code    C.int = C.SQLITE_ERROR
		message string
	)

	serr, ok := err.(*Error)
	if ok {
		code = C.int(serr.Code)
		message = serr.Message
	} else {
		message = err.Error()
	}

	*errmsg = sqliteString(message)

	return code
}

//export gosqliteConnect
func gosqliteConnect(conn *C.sqlite3, handle unsafe.Pointer, out *C.uintptr_t, errmsg **C.char) C.int {
	module := cgo.Handle(handle).Value().(EponymousModule)

	declaration := C.CString(module.Declaration())
	defer C.free(unsafe.Pointer(declaration))

	rv := C.sqlite3_declare_vtab(conn, declaration)
	if rv != C.SQLITE_OK {
		*errmsg = C.strdup(C.sqlite3_errmsg(conn))
		return rv
	}

	vtable, err := module.Connect()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	*out = C.uintptr_t(cgo.NewHandle(vtable))
	return C.SQLITE_OK
}

//export gosqliteCreate
func gosqliteCreate(conn *C.sqlite3, handle unsafe.Pointer, out *C.uintptr_t, errmsg **C.char) C.int {
	module := cgo.Handle(handle).Value().(Module)

	declaration := C.CString(module.Declaration())
	defer C.free(unsafe.Pointer(declaration))

	rv := C.sqlite3_declare_vtab(conn, declaration)
	if rv != C.SQLITE_OK {
		*errmsg = C.strdup(C.sqlite3_errmsg(conn))
		return rv
	}

	vtable, err := module.Create()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	*out = C.uintptr_t(cgo.NewHandle(vtable))
	return C.SQLITE_OK
}

//export gosqliteDisconnect
func gosqliteDisconnect(handle unsafe.Pointer, errmsg **C.char) C.int {
	h := cgo.Handle(handle)
	vtable := h.Value().(EponymousVirtualTable)

	err := vtable.Disconnect()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	h.Delete()

	return C.SQLITE_OK
}

//export gosqliteDestroy
func gosqliteDestroy(handle unsafe.Pointer, errmsg **C.char) C.int {
	h := cgo.Handle(handle)
	vtable := h.Value().(VirtualTable)

	err := vtable.Destroy()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	h.Delete()

	return C.SQLITE_OK
}

//export gosqliteBestIndex
func gosqliteBestIndex(handle unsafe.Pointer, out unsafe.Pointer, errmsg **C.char) C.int {
	vtable := cgo.Handle(handle).Value().(EponymousVirtualTable)
	info := (*C.sqlite3_index_info)(out)

	constraintsSlice := unsafe.Slice(info.aConstraint, int(info.nConstraint))
	constraints := make([]IndexConstraint, len(constraintsSlice))
	for i, constraint := range constraintsSlice {
		constraints[i] = IndexConstraint{
			Column:   int(constraint.iColumn),
			Operator: uint8(constraint.op),
			Usable:   constraint.usable != 0,
		}
	}

	orderBySlice := unsafe.Slice(info.aOrderBy, int(info.nOrderBy))
	orderBy := make([]IndexOrderBy, len(orderBySlice))
	for i, order := range orderBySlice {
		var direction = OrderByAsc

		if order.desc != 0 {
			direction = OrderByDesc
		}

		orderBy[i] = IndexOrderBy{
			Column:    int(order.iColumn),
			Direction: direction,
		}
	}

	result, err := vtable.BestIndex(constraints, orderBy)
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	info.idxNum = C.int(result.IndexNumber)
	info.idxStr = sqliteString(result.IndexName)
	info.needToFreeIdxStr = C.int(1)

	if result.OrderByConsumed {
		info.orderByConsumed = C.int(1)
	} else {
		info.orderByConsumed = C.int(0)
	}

	info.estimatedCost = C.double(result.EstimatedCost)
	info.estimatedRows = C.sqlite3_int64(result.EstimatedRows)

	return C.SQLITE_OK
}

//export gosqliteOpen
func gosqliteOpen(handle unsafe.Pointer, out *C.uintptr_t, errmsg **C.char) C.int {
	vtable := cgo.Handle(handle).Value().(EponymousVirtualTable)

	cursor, err := vtable.Open()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	*out = C.uintptr_t(cgo.NewHandle(cursor))

	return C.SQLITE_OK
}

//export gosqliteClose
func gosqliteClose(handle unsafe.Pointer, errmsg **C.char) C.int {
	cursor := cgo.Handle(handle).Value().(VirtualTableCursor)

	err := cursor.Close()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	return C.SQLITE_OK
}

//export gosqliteFilter
func gosqliteFilter(handle unsafe.Pointer, indexId C.int, indexName *C.char, argc C.int, argv **C.sqlite3_value, errmsg **C.char) C.int {
	cursor := cgo.Handle(handle).Value().(VirtualTableCursor)

	argvSlice := unsafe.Slice(argv, int(argc))
	values := make([]any, argc)
	for i, value := range argvSlice {
		switch C.sqlite3_value_type(value) {
		case C.SQLITE_INTEGER:
			values[i] = int64(C.sqlite3_value_int64(value))
		case C.SQLITE_FLOAT:
			values[i] = float64(C.sqlite3_value_double(value))
		case C.SQLITE_TEXT:
			values[i] = C.GoStringN(
				(*C.char)(unsafe.Pointer(C.sqlite3_value_text(value))),
				C.sqlite3_value_bytes(value))
		case C.SQLITE_BLOB:
			values[i] = C.GoBytes(C.sqlite3_value_blob(value), C.sqlite3_value_bytes(value))
		default:
			values[i] = nil
		}
	}

	err := cursor.Filter(int(indexId), C.GoString(indexName), values)
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	return C.SQLITE_OK
}

//export gosqliteNext
func gosqliteNext(handle unsafe.Pointer, errmsg **C.char) C.int {
	cursor := cgo.Handle(handle).Value().(VirtualTableCursor)

	err := cursor.Next()
	if err != nil {
		return gosqliteError(err, errmsg)
	}

	return C.SQLITE_OK
}

//export gosqliteEOF
func gosqliteEOF(handle unsafe.Pointer) C.int {
	cursor := cgo.Handle(handle).Value().(VirtualTableCursor)

	if cursor.EOF() {
		return C.int(1)
	}

	return C.int(0)
}

//export gosqliteColumn
func gosqliteColumn(column int) C.int {
	return C.SQLITE_OK
}

//export gosqliteRowid
func gosqliteRowid() C.int {
	return C.SQLITE_OK
}

func (c *Conn) CreateModule(name string, module EponymousModule) error {
	namePtr := C.CString(name)
	defer C.free(unsafe.Pointer(namePtr))

	handle := cgo.NewHandle(module)

	var rv C.int
	switch module.(type) {
	case Module:
		rv = C.gosqlite_create_module(c.conn, namePtr, C.uintptr_t(handle))

	default:
		rv = C.gosqlite_create_eponymous_module(c.conn, namePtr, C.uintptr_t(handle))
	}

	if rv != C.SQLITE_OK {
		handle.Delete()
		return c.error(int(rv))
	}

	c.modules = append(c.modules, handle)

	return nil
}
