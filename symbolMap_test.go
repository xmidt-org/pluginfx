package pluginfx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SymbolMapSuite struct {
	suite.Suite
}

func (suite *SymbolMapSuite) TestSet() {
	suite.Run("UntypedNil", func() {
		var sm SymbolMap
		suite.Panics(func() {
			sm.Set("foo", nil)
		})
	})

	suite.Run("Nil", func() {
		var (
			sm    SymbolMap
			value *int
		)

		suite.Panics(func() {
			sm.Set("foo", value)
		})
	})

	suite.Run("Function", func() {
		var (
			sm SymbolMap

			called bool
			value  = func() {
				called = true
			}
		)

		sm.Set("foo", value)
		f, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(f)

		f.(func())()
		suite.True(called)
	})

	suite.Run("Value", func() {
		var sm SymbolMap
		sm.Set("foo", 123)

		v, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(123, *v.(*int))
	})

	suite.Run("Pointer", func() {
		var (
			sm       SymbolMap
			expected = 123
		)

		sm.Set("foo", &expected)

		v, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(expected, *v.(*int))
	})
}

func (suite *SymbolMapSuite) TestDel() {
	suite.Run("Empty", func() {
		var sm SymbolMap
		sm.Del("foo")
	})

	suite.Run("NotPresent", func() {
		var sm SymbolMap
		sm.Set("foo", 123)
		sm.Del("bar")

		v, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(123, *v.(*int))
	})

	suite.Run("Present", func() {
		var sm SymbolMap
		sm.Set("foo", 123)
		sm.Del("foo")

		v, err := sm.Lookup("foo")
		suite.Nil(v)
		var mse *MissingSymbolError
		suite.True(errors.As(err, &mse))
	})
}

func (suite *SymbolMapSuite) TestLookup() {
	suite.Run("Empty", func() {
		var sm SymbolMap

		v, err := sm.Lookup("foo")
		suite.Nil(v)
		var mse *MissingSymbolError
		suite.True(errors.As(err, &mse))
	})

	suite.Run("Found", func() {
		var sm SymbolMap
		sm.Set("foo", 123)

		v, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(123, *v.(*int))
	})

	suite.Run("NotFound", func() {
		var sm SymbolMap
		sm.Set("foo", 123)

		v, err := sm.Lookup("bar")
		suite.Nil(v)
		var mse *MissingSymbolError
		suite.True(errors.As(err, &mse))
	})
}

func (suite *SymbolMapSuite) TestNewSymbolMap() {
	suite.Run("Nil", func() {
		sm := NewSymbolMap(nil)
		suite.Require().NotNil(sm)

		v, err := sm.Lookup("nosuch")
		suite.Nil(v)

		var mse *MissingSymbolError
		suite.True(errors.As(err, &mse))
	})

	suite.Run("NotNil", func() {
		var (
			called bool
			sm     = NewSymbolMap(map[string]interface{}{
				"foo": 123,
				"bar": func() {
					called = true
				},
			})
		)

		suite.Require().NotNil(sm)

		v, err := sm.Lookup("foo")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)
		suite.Equal(123, *v.(*int))

		v, err = sm.Lookup("bar")
		suite.Require().NoError(err)
		suite.Require().NotNil(v)

		v.(func())()
		suite.True(called)

		v, err = sm.Lookup("nosuch")
		suite.Nil(v)
		var mse *MissingSymbolError
		suite.True(errors.As(err, &mse))
	})
}

func TestSymbolMap(t *testing.T) {
	suite.Run(t, new(SymbolMapSuite))
}
