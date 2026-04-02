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

func TestModuleCreation(t *testing.T) {
	t.Parallel()

	virtualTable := new(MockVirtualTable)
	module := new(MockModule)
	conn, err := sqlite.Open(":memory:")
	require.NoError(t, err)

	module.On("Declaration").Return("CREATE TABLE mock(id INTEGER)").Once()
	module.On("Connect").Return(virtualTable, nil).Once()

	err = conn.CreateModule("mock", module)
	require.NoError(t, err)

	statement, err := conn.Prepare("CREATE VIRTUAL TABLE mock USING mock")
	require.NoError(t, err)

	for {
		done, err := statement.Step()
		require.NoError(t, err)
		if done {
			break
		}
	}
	require.NoError(t, statement.Finalize())

	module.AssertExpectations(t)
}
