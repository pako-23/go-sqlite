package sqlite

import "C"

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

type VirtualTable interface {
	BestIndex(constraints []IndexConstraint, order []IndexOrderBy) error
	Disconnect() error
	Open() (VirtualTableCursor, error)
}

type Module interface {
	Connect() (VirtualTable, error)
}

func (c *Conn) CreateModule(name string, module Module) error {
	return nil
}

func (c *Conn) DeclareVirtualTable(query string) error {
	return nil
}
