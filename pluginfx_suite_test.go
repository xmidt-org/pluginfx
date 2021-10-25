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
	return mse
}
