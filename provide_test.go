package pluginfx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const expectedNewValue float64 = 67.5

type ProvideSuite struct {
	suite.Suite
}

func (suite *ProvideSuite) testPGlobal() {
	var (
		value float64
		p     Plugin

		app = fxtest.New(
			suite.T(),
			P{
				Path: "./sample.so",
				Symbols: Symbols{
					Names: []interface{}{
						"New",
					},
				},
				Lifecycle: Lifecycle{
					OnStart: "Initialize",
					OnStop:  "Shutdown",
				},
			}.Provide(),
			fx.Populate(&value, &p),
		)
	)

	app.RequireStart()
	app.RequireStop()

	suite.Equal(expectedNewValue, value)
	suite.NotNil(p)
}

func (suite *ProvideSuite) testPAnonymous() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			P{
				Anonymous: true,
				Name:      "ShouldBeIgnored",
				Path:      "./sample.so",
				Symbols: Symbols{
					Names: []interface{}{
						"New",
					},
				},
				Lifecycle: Lifecycle{
					OnStart: "Initialize",
					OnStop:  "Shutdown",
				},
			}.Provide(),
			fx.Populate(&value),
			fx.Invoke(
				func(in struct {
					fx.In
					Plugin Plugin `optional:"true"`
				}) {
					suite.Nil(in.Plugin)
				},
			),
		)
	)

	app.RequireStart()
	app.RequireStop()

	suite.Equal(expectedNewValue, value)
}

func (suite *ProvideSuite) testPNamed() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			P{
				Name: "MyPlugin",
				Path: "./sample.so",
				Symbols: Symbols{
					Names: []interface{}{
						"New",
					},
				},
				Lifecycle: Lifecycle{
					OnStart: "Initialize",
					OnStop:  "Shutdown",
				},
			}.Provide(),
			fx.Populate(&value),
			fx.Invoke(
				func(in struct {
					fx.In
					Plugin Plugin `name:"MyPlugin"`
				}) {
				},
			),
		)
	)

	app.RequireStart()
	app.RequireStop()

	suite.Equal(expectedNewValue, value)
}

func (suite *ProvideSuite) testPGroup() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			P{
				Group: "plugins",
				Path:  "./sample.so",
				Symbols: Symbols{
					Names: []interface{}{
						"New",
					},
				},
				Lifecycle: Lifecycle{
					OnStart: "Initialize",
					OnStop:  "Shutdown",
				},
			}.Provide(),
			fx.Populate(&value),
			fx.Invoke(
				func(in struct {
					fx.In
					Plugins []Plugin `group:"plugins"`
				}) {
					suite.Len(in.Plugins, 1)
				},
			),
		)
	)

	app.RequireStart()
	app.RequireStop()

	suite.Equal(expectedNewValue, value)
}

func (suite *ProvideSuite) testPAnonymousError() {
	app := fx.New(
		P{
			// Even when anonymous, a plugin that can't be loaded should
			// still short-circuit with an error.
			Anonymous: true,
			Path:      "/no/such/plugin.123",
		}.Provide(),
	)

	err := app.Err()
	suite.Require().Error(err)

	var oe *OpenError
	suite.Require().True(errors.As(err, &oe))
	suite.NotEmpty(oe.Error())
}

func (suite *ProvideSuite) TestP() {
	suite.Run("Global", suite.testPGlobal)
	suite.Run("Anonymous", suite.testPAnonymous)
	suite.Run("Named", suite.testPNamed)
	suite.Run("Group", suite.testPGroup)
	suite.Run("AnonymousError", suite.testPAnonymousError)
}

func TestProvide(t *testing.T) {
	suite.Run(t, new(ProvideSuite))
}
