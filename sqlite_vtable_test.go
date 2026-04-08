package sqlite_test

import (
	"fmt"
	"testing"

	"github.com/pako-23/go-sqlite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockModule struct {
	mock.Mock
}

func (m *MockModule) Declaration() string {
	args := m.Called()

	return args.Get(0).(string)
}

func (m *MockModule) Connect() (sqlite.EponymousVirtualTable, error) {
	args := m.Called()

	return args.Get(0).(sqlite.EponymousVirtualTable), args.Error(1)
}

func (m *MockModule) Create() (sqlite.VirtualTable, error) {
	args := m.Called()

	return args.Get(0).(sqlite.VirtualTable), args.Error(1)
}

type MockVirtualTable struct {
	mock.Mock
}

func (m *MockVirtualTable) BestIndex(constraints []sqlite.IndexConstraint, order []sqlite.IndexOrderBy) (sqlite.IndexResult, error) {
	args := m.Called(constraints, order)

	return args.Get(0).(sqlite.IndexResult), args.Error(1)
}

func (m *MockVirtualTable) Disconnect() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockVirtualTable) Destroy() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockVirtualTable) Open() (sqlite.VirtualTableCursor, error) {
	args := m.Called()

	return args.Get(0).(sqlite.VirtualTableCursor), args.Error(1)
}

func (m *MockVirtualTable) Delete(id any) error {
	args := m.Called(id)

	return args.Error(0)
}

func (m *MockVirtualTable) Insert(id any, values []any) (int64, error) {
	args := m.Called(id, values)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVirtualTable) Update(id any, values []any, newId ...any) error {
	args := m.Called(id, values, newId)

	return args.Error(0)
}

type MockCursor struct {
	mock.Mock
}

func (m *MockCursor) Close() error { return m.Called().Error(0) }

func (m *MockCursor) Filter(indexId int, indexName string, values []any) error {
	return m.Called(indexId, indexName, values).Error(0)
}

func (m *MockCursor) Next() error { return m.Called().Error(0) }

func (m *MockCursor) EOF() bool { return m.Called().Bool(0) }

func (m *MockCursor) Rowid() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCursor) Column(column int) (any, error) {
	args := m.Called(column)
	return args.Get(0), args.Error(1)
}

func queryExec(conn *sqlite.Conn, query string) error {
	statement, err := conn.Prepare(query)
	if err != nil {
		return err
	}

	for {
		done, err := statement.Step()
		if err != nil {
			return err
		}
		if done {
			break
		}
	}

	return statement.Finalize()
}

func TestModuleCreate(t *testing.T) {
	t.Parallel()

	vtable := new(MockVirtualTable)
	module := new(MockModule)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE mock(id INTEGER)").Once()
	module.On("Connect").Return(vtable, nil).Once()
	vtable.On("Disconnect").Return(nil).Once()

	err = conn.CreateModule("mock", module)
	require.NoError(t, err)

	err = queryExec(conn, "CREATE VIRTUAL TABLE mock USING mock")
	require.NoError(t, err)

	require.NoError(t, conn.Close())
	module.AssertExpectations(t)
	vtable.AssertExpectations(t)
}

func TestModuleDestroy(t *testing.T) {
	t.Parallel()

	times := 3
	vtable := new(MockVirtualTable)
	module := new(MockModule)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE mock(id INTEGER)").Times(times)
	module.On("Connect").Return(vtable, nil).Times(times)
	vtable.On("Destroy").Return(nil).Times(times)
	err = conn.CreateModule("mock", module)
	require.NoError(t, err)

	for range times {
		err = queryExec(conn, "CREATE VIRTUAL TABLE mock USING mock")
		require.NoError(t, err)

		err = queryExec(conn, "DROP TABLE mock")
		require.NoError(t, err)
	}

	require.NoError(t, conn.Close())
	module.AssertExpectations(t)
	vtable.AssertExpectations(t)
}

func TestVirtualTableIteration(t *testing.T) {
	t.Parallel()

	vtable := new(MockVirtualTable)
	cursor := new(MockCursor)
	module := new(MockModule)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE test(a INTEGER, b REAL, c TEXT, d BLOB, e INTEGER)")
	module.On("Connect").Return(vtable, nil).Once()
	vtable.On("Disconnect").Return(nil).Once()
	vtable.On("Open").Return(cursor, nil).Once()
	vtable.On("BestIndex", mock.Anything, mock.Anything).Return(sqlite.IndexResult{}, nil).Once()

	cursor.On("Filter", 0, "", mock.Anything).Return(nil).Once()
	cursor.On("EOF").Return(false).Twice()
	cursor.On("EOF").Return(true).Once()

	cursor.On("Rowid").Return(int64(0), nil).Once()
	cursor.On("Rowid").Return(int64(1), nil).Once()
	cursor.On("Column", 0).Return(int64(42), nil).Twice()

	cursor.On("Column", 1).Return(3.14, nil).Twice()
	cursor.On("Column", 2).Return("hello", nil).Twice()
	cursor.On("Column", 3).Return([]byte{0xFF, 0x00}, nil).Once()
	cursor.On("Column", 3).Return([]byte{}, nil).Once()
	cursor.On("Column", 4).Return(nil, nil).Once()
	cursor.On("Column", 4).Return(int64(10), nil).Once()
	cursor.On("Next").Return(nil).Twice()
	cursor.On("Close").Return(nil).Once()

	require.NoError(t, conn.CreateModule("test_mod", module))
	require.NoError(t, queryExec(conn, "CREATE VIRTUAL TABLE t1 USING test_mod"))

	statement, err := conn.Prepare("SELECT ROWID, * FROM t1")
	require.NoError(t, err)
	require.Equal(t, 6, statement.ColumnCount())

	names := make([]string, statement.ColumnCount())
	for i := range names {
		name, err := statement.ColumnName(i)
		require.NoError(t, err)
		names[i] = name
	}

	require.Equal(t, []string{"rowid", "a", "b", "c", "d", "e"}, names)

	done, err := statement.Step()
	require.NoError(t, err)
	require.False(t, done)

	row, err := statement.Row()
	require.NoError(t, err)
	require.Equal(t, int64(0), row[0])
	require.Equal(t, int64(42), row[1])
	require.Equal(t, 3.14, row[2])
	require.Equal(t, "hello", row[3])
	require.Equal(t, []byte{0xFF, 0x00}, row[4])
	require.Nil(t, row[5])

	done, err = statement.Step()
	require.NoError(t, err)
	require.False(t, done)

	row, err = statement.Row()
	require.NoError(t, err)
	require.Equal(t, int64(1), row[0])
	require.Equal(t, int64(42), row[1])
	require.Equal(t, 3.14, row[2])
	require.Equal(t, "hello", row[3])
	require.Equal(t, nil, row[4])
	require.Equal(t, int64(10), row[5])

	done, err = statement.Step()
	require.NoError(t, err)
	require.True(t, done)

	require.NoError(t, statement.Finalize())
	require.NoError(t, conn.Close())

	module.AssertExpectations(t)
	vtable.AssertExpectations(t)
	cursor.AssertExpectations(t)
}

func TestVirtualTableUpdates(t *testing.T) {
	t.Parallel()

	vtable := new(MockVirtualTable)
	module := new(MockModule)
	cursor := new(MockCursor)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE test(a INTEGER PRIMARY KEY)")
	module.On("Connect").Return(vtable, nil).Once()
	vtable.On("Disconnect").Return(nil).Once()

	require.NoError(t, conn.CreateModule("test_mod", module))
	require.NoError(t, queryExec(conn, "CREATE VIRTUAL TABLE t1 USING test_mod"))

	t.Run("insertion", func(t *testing.T) {
		for i := 1; i < 4; i++ {
			vtable.On("Insert", nil, []any{int64(i)}).Return(int64(i), nil).Once()
		}

		err = queryExec(conn, "INSERT INTO t1(a) VALUES (1), (2), (3)")
		require.NoError(t, err)
	})

	t.Run("deletion rowid existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).Return(sqlite.IndexResult{}, nil).Once()
		cursor.On("Filter", 0, "", mock.Anything).Return(nil).Once()
		cursor.On("EOF").Return(false).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Close").Return(nil).Once()
		cursor.On("Rowid").Return(int64(1), nil).Twice()
		cursor.On("Next").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()
		vtable.On("Delete", int64(1)).Return(nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE ROWID = 1")
		require.NoError(t, err)
	})

	t.Run("deletion rowid not existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).Return(sqlite.IndexResult{}, nil).Once()
		cursor.On("Filter", 0, "", mock.Anything).Return(nil).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Close").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE ROWID = 1")
		require.NoError(t, err)
	})

	t.Run("deletion key existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).
			Return(sqlite.IndexResult{
				IndexNumber: 0,
				IndexName:   "a-column",
				Usage:       []bool{true},
			}, nil).Once()
		cursor.On("Filter", 0, "a-column", []any{int64(1)}).Return(nil).Once()
		cursor.On("EOF").Return(false).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Close").Return(nil).Once()
		cursor.On("Rowid").Return(int64(0), nil)
		cursor.On("Next").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()
		vtable.On("Delete", int64(0)).Return(nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE a = 1")
		require.NoError(t, err)
	})

	t.Run("deletion key not existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).
			Return(sqlite.IndexResult{
				IndexNumber: 0,
				IndexName:   "a-column",
				Usage:       []bool{true},
			}, nil).Once()
		cursor.On("Filter", 0, "a-column", []any{int64(1)}).Return(nil).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Close").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE a = 1")
		require.NoError(t, err)
	})

	require.NoError(t, conn.Close())
	cursor.AssertExpectations(t)
	vtable.AssertExpectations(t)
	module.AssertExpectations(t)
}

func TestVirtualTableUpdatesNoRowid(t *testing.T) {
	t.Parallel()

	vtable := new(MockVirtualTable)
	module := new(MockModule)
	cursor := new(MockCursor)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE test(a TEXT PRIMARY KEY) WITHOUT ROWID")
	module.On("Connect").Return(vtable, nil).Once()
	vtable.On("Disconnect").Return(nil).Once()

	require.NoError(t, conn.CreateModule("test_mod", module))
	require.NoError(t, queryExec(conn, "CREATE VIRTUAL TABLE t1 USING test_mod"))

	t.Run("insertion", func(t *testing.T) {
		for i := 1; i < 4; i++ {
			vtable.On("Insert", nil, []any{fmt.Sprintf("%d", i)}).Return(int64(0), nil).Once()
		}

		err = queryExec(conn, "INSERT INTO t1(a) VALUES ('1'), ('2'), ('3')")
		require.NoError(t, err)
	})

	t.Run("deletion existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).
			Return(sqlite.IndexResult{
				IndexNumber: 0,
				IndexName:   "a-column",
				Usage:       []bool{true},
			}, nil).Once()
		cursor.On("Filter", 0, "a-column", []any{"1"}).Return(nil).Once()
		cursor.On("EOF").Return(false).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Column", 0).Return(int64(1), nil).Once()
		cursor.On("Close").Return(nil).Once()
		cursor.On("Next").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()
		vtable.On("Delete", "1").Return(nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE a = '1'")
		require.NoError(t, err)
	})

	t.Run("deletion not existing", func(t *testing.T) {
		vtable.On("BestIndex", mock.Anything, mock.Anything).
			Return(sqlite.IndexResult{
				IndexNumber: 0,
				IndexName:   "a-column",
				Usage:       []bool{true},
			}, nil).Once()
		cursor.On("Filter", 0, "a-column", []any{"1"}).Return(nil).Once()
		cursor.On("EOF").Return(true).Once()
		cursor.On("Close").Return(nil).Once()
		vtable.On("Open").Return(cursor, nil).Once()

		err = queryExec(conn, "DELETE FROM t1 WHERE a = '1'")
		require.NoError(t, err)
	})

	require.NoError(t, conn.Close())
	cursor.AssertExpectations(t)
	vtable.AssertExpectations(t)
	module.AssertExpectations(t)
}
