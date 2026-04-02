package sqlite

/*
#include "gosqlite_vtable.h"
#include <stdlib.h>
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

type IndexConstraint struct {
	Column   int
	Operator uint8
	Usable   bool
}

type IndexOrderBy struct {
	Column    int
	Direction uint8
}

type EponymousVirtualTable interface {
	BestIndex(constraints []IndexConstraint, order []IndexOrderBy) error
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

//export gosqliteConnectImpl
func gosqliteConnectImpl(conn *C.sqlite3, handle unsafe.Pointer, errorMessage **C.char) C.uintptr_t {
	module := cgo.Handle(handle).Value().(EponymousModule)

	declaration := C.CString(module.Declaration())
	defer C.free(unsafe.Pointer(declaration))

	rv := C.sqlite3_declare_vtab(conn, declaration)
	if rv != C.SQLITE_OK {
		// TODO implement real error handling
		return 0
	}

	vtable, err := module.Connect()
	if err != nil {
		// TODO copy error message into errorMessage
		return 0
	}

	return C.uintptr_t(cgo.NewHandle(vtable))
}

//export gosqliteCreateImpl
func gosqliteCreateImpl(conn *C.sqlite3, handle unsafe.Pointer, errorMessage **C.char) C.uintptr_t {
	module := cgo.Handle(handle).Value().(Module)

	declaration := C.CString(module.Declaration())
	defer C.free(unsafe.Pointer(declaration))

	rv := C.sqlite3_declare_vtab(conn, declaration)
	if rv != C.SQLITE_OK {
		// TODO implement real error handling
		return 0
	}

	vtable, err := module.Create()
	if err != nil {
		// TODO copy error message into errorMessage
		return 0
	}

	return C.uintptr_t(cgo.NewHandle(vtable))
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

	return nil
}
