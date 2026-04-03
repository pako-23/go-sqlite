package sqlite_test

import (
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

func (m *MockVirtualTable) BestIndex(constraints []sqlite.IndexConstraint, order []sqlite.IndexOrderBy) error {
	args := m.Called(constraints, order)

	return args.Error(0)
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
