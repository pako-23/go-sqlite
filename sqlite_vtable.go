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

	*errmsg = C.CString(message)

	return code
}

//export gosqliteConnectImpl
func gosqliteConnectImpl(conn *C.sqlite3, handle unsafe.Pointer, out *C.uintptr_t, errmsg **C.char) C.int {
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

//export gosqliteCreateImpl
func gosqliteCreateImpl(conn *C.sqlite3, handle unsafe.Pointer, out *C.uintptr_t, errmsg **C.char) C.int {
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

//export gosqliteBestIndexImpl
func gosqliteBestIndexImpl() {
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

//export gosqliteOpen
func gosqliteOpen() {
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
