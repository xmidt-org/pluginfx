package pluginfx

import (
	"errors"

	"github.com/stretchr/testify/suite"
)

type PluginfxSuite struct {
	suite.Suite
}

func (suite *PluginfxSuite) missingSymbolError(expectedName string, err error) *MissingSymbolError {
	var mse *MissingSymbolError
	suite.Require().True(errors.As(err, &mse))
	suite.Equal(expectedName, mse.Name)
	suite.Equal(mse.Err, errors.Unwrap(mse))
	suite.NotEmpty(mse.Error())

	return mse
}

func (suite *PluginfxSuite) openError(expectedPath string, err error) *OpenError {
	var oe *OpenError
	suite.Require().True(errors.As(err, &oe))
	suite.Equal(expectedPath, oe.Path)
	suite.Error(oe.Err)
	suite.Equal(oe.Err, errors.Unwrap(oe))
	suite.NotEmpty(oe.Error())

	return oe
}

func (suite *PluginfxSuite) openSuccess(p Plugin, err error) Plugin {
	suite.NoError(err)
	suite.Require().NotNil(p)
	return p
}
