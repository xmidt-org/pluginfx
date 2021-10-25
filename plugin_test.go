package pluginfx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PluginSuite struct {
	PluginfxSuite
}

func (suite *PluginSuite) TestOpen() {
	suite.Run("Nosuch", func() {
		p, err := Open("nosuch")
		suite.Nil(p)
		suite.openError("nosuch", err)
	})

	suite.Run("Sample", func() {
		p := suite.openSuccess(Open(samplePath))

		v, err := p.Lookup("Value")
		suite.NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(12, *v.(*int))
	})
}

func (suite *PluginSuite) TestLookup() {
	suite.Run("SymbolMap", func() {
		suite.Run("Found", func() {
			var sm SymbolMap
			sm.Set("Foo", 12)

			v, err := Lookup(&sm, "Foo")
			suite.NoError(err)
			suite.Require().NotNil(v)
			suite.Equal(12, *v.(*int))
		})

		suite.Run("Missing", func() {
			var sm SymbolMap

			v, err := Lookup(&sm, "Foo")
			suite.Nil(v)
			mse := suite.missingSymbolError("Foo", err)
			suite.NoError(mse.Err)
		})
	})

	suite.Run("Sample", func() {
		suite.Run("Found", func() {
			p := suite.openSuccess(Open(samplePath))

			v, err := Lookup(p, "Value")
			suite.NoError(err)
			suite.Require().NotNil(v)
			suite.Equal(12, *v.(*int))
		})

		suite.Run("Missing", func() {
			p := suite.openSuccess(Open(samplePath))

			v, err := Lookup(p, "Nosuch")
			suite.Nil(v)
			mse := suite.missingSymbolError("Nosuch", err)
			suite.Error(mse.Err)
		})
	})
}

func (suite *PluginSuite) TestIsMissingSymbolError() {
	suite.Run("Nil", func() {
		suite.False(IsMissingSymbolError(nil))
	})

	suite.Run("DifferentError", func() {
		suite.False(
			IsMissingSymbolError(errors.New("some error")),
		)
	})

	suite.Run("Missing", func() {
		suite.True(
			IsMissingSymbolError(new(MissingSymbolError)),
		)
	})
}

func TestPlugin(t *testing.T) {
	suite.Run(t, new(PluginSuite))
}
