package pluginfx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PluginSuite struct {
	suite.Suite
}

func (suite *PluginSuite) TestOpen() {
	suite.Run("Nosuch", func() {
	})
}

func TestPlugin(t *testing.T) {
	suite.Run(t, new(PluginSuite))
}
