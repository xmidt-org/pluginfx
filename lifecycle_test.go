package pluginfx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
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

	suite.Run("Missing", func() {
		var (
			m = new(mockLifecycleMethod)

			lifecycle = Lifecycle{
				OnStart: "NoSuch",
			}

			sm  SymbolMap
			app = fx.New(
				lifecycle.Bind(&sm),
			)
		)

		suite.Error(app.Err())
		m.AssertExpectations(suite.T())
	})

	suite.Run("Invalid", func() {
		var (
			m = new(mockLifecycleMethod)

			lifecycle = Lifecycle{
				OnStart: "Invalid",
			}

			app = fx.New(
				lifecycle.Bind(NewSymbolMap(map[string]interface{}{
					"Invalid": func(int) bool { return true },
				})),
			)
		)

		err := app.Err()
		suite.Require().Error(err)
		var ile *InvalidLifecycleError
		suite.Require().True(errors.As(err, &ile))
		suite.Require().NotNil(ile)
		suite.NotEmpty(ile.Error())

		m.AssertExpectations(suite.T())
	})
}

func (suite *LifecycleSuite) testOnStopSuccess(onStop interface{}) {
	var (
		lifecycle = Lifecycle{
			OnStop: "TestOnStop",
		}

		app = fxtest.New(
			suite.T(),
			lifecycle.Bind(NewSymbolMap(map[string]interface{}{
				lifecycle.OnStop: onStop,
			})),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *LifecycleSuite) TestOnStop() {
	suite.Run("Success", func() {
		suite.Run("Simple", func() {
			m := new(mockLifecycleMethod)
			m.ExpectSimple()
			suite.testOnStopSuccess(m.Simple)
			m.AssertExpectations(suite.T())
		})

		suite.Run("ReturnsError", func() {
			m := new(mockLifecycleMethod)
			m.ExpectReturnsError(nil)
			suite.testOnStopSuccess(m.ReturnsError)
			m.AssertExpectations(suite.T())
		})

		suite.Run("AcceptsContext", func() {
			m := new(mockLifecycleMethod)
			m.ExpectAcceptsContext(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }))
			suite.testOnStopSuccess(m.AcceptsContext)
			m.AssertExpectations(suite.T())
		})

		suite.Run("Full", func() {
			m := new(mockLifecycleMethod)
			m.ExpectFull(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), nil)
			suite.testOnStopSuccess(m.Full)
			m.AssertExpectations(suite.T())
		})
	})

	suite.Run("Missing", func() {
		var (
			m = new(mockLifecycleMethod)

			lifecycle = Lifecycle{
				OnStop: "NoSuch",
			}

			sm  SymbolMap
			app = fx.New(
				lifecycle.Bind(&sm),
			)
		)

		suite.Error(app.Err())
		m.AssertExpectations(suite.T())
	})

	suite.Run("Invalid", func() {
		var (
			m = new(mockLifecycleMethod)

			lifecycle = Lifecycle{
				OnStop: "Invalid",
			}

			app = fx.New(
				lifecycle.Bind(NewSymbolMap(map[string]interface{}{
					"Invalid": func(int) bool { return true },
				})),
			)
		)

		err := app.Err()
		suite.Require().Error(err)
		var ile *InvalidLifecycleError
		suite.Require().True(errors.As(err, &ile))
		suite.Require().NotNil(ile)
		suite.NotEmpty(ile.Error())

		m.AssertExpectations(suite.T())
	})
}

func TestLifecycle(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}
