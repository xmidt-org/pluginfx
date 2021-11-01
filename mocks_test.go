package pluginfx

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockLifecycleMethod struct {
	mock.Mock
}

func (m *mockLifecycleMethod) Simple() {
	m.Called()
}

func (m *mockLifecycleMethod) ExpectSimple() *mock.Call {
	return m.On("Simple")
}

func (m *mockLifecycleMethod) ReturnsError() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLifecycleMethod) ExpectReturnsError(err error) *mock.Call {
	return m.On("ReturnsError").Return(err)
}

func (m *mockLifecycleMethod) AcceptsContext(ctx context.Context) {
	m.Called(ctx)
}

func (m *mockLifecycleMethod) ExpectAcceptsContext(ctx interface{}) *mock.Call {
	return m.On("AcceptsContext", ctx)
}

func (m *mockLifecycleMethod) Full(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockLifecycleMethod) ExpectFull(ctx interface{}, err error) *mock.Call {
	return m.On("Full", ctx).Return(err)
}
