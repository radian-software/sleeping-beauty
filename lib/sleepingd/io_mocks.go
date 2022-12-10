package sleepingd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConn struct {
	mock.Mock
	t *testing.T
}

func newMockConn(t *testing.T) *mockConn {
	return &mockConn{mock.Mock{}, t}
}

func (m *mockConn) Read(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConn) ReturnFromNextRead(b []byte, notify chan<- struct{}) {
	m.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buf := args.Get(0).([]byte)
		assert.GreaterOrEqual(m.t, len(buf), len(b))
		copy(buf, b)
		if notify != nil {
			notify <- struct{}{}
		}
	}).Return(len(b), nil).Once()
}

func (m *mockConn) ExpectInNextWrite(b []byte, notify chan<- struct{}) {
	m.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		args.Assert(m.t, b)
		if notify != nil {
			notify <- struct{}{}
		}
	}).Return(len(b), nil).Once()
}
