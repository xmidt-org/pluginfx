package pluginfx

import (
	"testing"

	"go.uber.org/fx"
)

func TestIt(t *testing.T) {
	app := fx.New(
		fx.Provide(
			func() (error, int) { return nil, 1 },
		),
	)

	t.Log(app.Err())
}
