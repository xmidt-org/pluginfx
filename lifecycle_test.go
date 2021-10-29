package pluginfx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"
)

type LifecycleSuite struct {
	PluginfxSuite
}

func (suite *LifecycleSuite) TestIgnoreMissing() {
	var (
		lifecycle = Lifecycle{
			OnStart:       "NosuchOnStart",
			OnStop:        "NosuchOnStop",
			IgnoreMissing: true,
		}

		app = fxtest.New(
			suite.T(),
			lifecycle.Bind(NewSymbolMap(nil)),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *LifecycleSuite) testOnStartSuccess(onStart interface{}) {
	var (
		lifecycle = Lifecycle{
			OnStart: "TestOnStart",
		}

		app = fxtest.New(
			suite.T(),
			lifecycle.Bind(NewSymbolMap(map[string]interface{}{
				lifecycle.OnStart: onStart,
			})),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *LifecycleSuite) TestOnStart() {
	suite.Run("Success", func() {
		suite.Run("Simple", func() {
			m := new(mockLifecycleMethod)
			m.ExpectSimple()
			suite.testOnStartSuccess(m.Simple)
			m.AssertExpectations(suite.T())
		})

		suite.Run("ReturnsError", func() {
			m := new(mockLifecycleMethod)
			m.ExpectReturnsError(nil)
			suite.testOnStartSuccess(m.ReturnsError)
			m.AssertExpectations(suite.T())
		})

		suite.Run("AcceptsContext", func() {
			m := new(mockLifecycleMethod)
			m.ExpectAcceptsContext(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }))
			suite.testOnStartSuccess(m.AcceptsContext)
			m.AssertExpectations(suite.T())
		})

		suite.Run("Full", func() {
			m := new(mockLifecycleMethod)
			m.ExpectFull(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), nil)
			suite.testOnStartSuccess(m.Full)
			m.AssertExpectations(suite.T())
		})
	})
}

func TestLifecycle(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
