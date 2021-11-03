package pluginfx

import (
	"errors"
	"path/filepath"
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
				Path: samplePath,
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

func (suite *ProvideSuite) testPExpandEnv() {
	var (
		value float64
		p     Plugin

		app = fxtest.New(
			suite.T(),
			P{
				Path: "${PWD}/" + samplePath,
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
				Path:      samplePath,
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
				Path: samplePath,
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
				Path:  samplePath,
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
	suite.Run("ExpandEnv", suite.testPExpandEnv)
	suite.Run("Anonymous", suite.testPAnonymous)
	suite.Run("Named", suite.testPNamed)
	suite.Run("Group", suite.testPGroup)
	suite.Run("AnonymousError", suite.testPAnonymousError)
}

func (suite *ProvideSuite) testSAnonymous() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			S{
				Paths: []string{samplePath},
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
		)
	)

	app.RequireStart()
	app.RequireStop()

	suite.Equal(expectedNewValue, value)
}

func (suite *ProvideSuite) testSGroup() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			S{
				Group: "plugins",
				Paths: []string{samplePath},
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

func (suite *ProvideSuite) testSExpandEnv() {
	var (
		value float64

		app = fxtest.New(
			suite.T(),
			S{
				Group: "plugins",
				Paths: []string{"${PWD}/*.so"},
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

func (suite *ProvideSuite) testSBadGlob() {
	app := fx.New(
		S{
			Paths: []string{"["},
		}.Provide(),
	)

	err := app.Err()
	suite.Require().Error(err)
	suite.True(errors.Is(err, filepath.ErrBadPattern))
}

func (suite *ProvideSuite) TestS() {
	suite.Run("Anonymous", suite.testSAnonymous)
	suite.Run("Group", suite.testSGroup)
	suite.Run("ExpandEnv", suite.testSExpandEnv)
	suite.Run("BadGlob", suite.testSBadGlob)
}

func TestProvide(t *testing.T) {
	suite.Run(t, new(ProvideSuite))
}
