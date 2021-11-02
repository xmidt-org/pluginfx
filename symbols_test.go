package pluginfx

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type SymbolsSuite struct {
	PluginfxSuite
}

func (suite *SymbolsSuite) testLoadSuccess() {
	var (
		invoke1Called bool
		invoke2Called bool
		invoke3Called bool

		sm = NewSymbols(
			"Constructor1", func() *bytes.Buffer {
				return new(bytes.Buffer)
			},
			"Constructor2", func(*bytes.Buffer) (*testing.T, error) {
				return suite.T(), nil
			},
			"Target1", func() *bytes.Buffer {
				return new(bytes.Buffer)
			},
			"Target2", func() *bytes.Buffer {
				return new(bytes.Buffer)
			},
			"Invoke1", func() {
				invoke1Called = true
			},
			"Invoke2", func() error {
				invoke2Called = true
				return nil
			},
			"Invoke3", func(struct {
				fx.In
				C1 *bytes.Buffer
				C2 *testing.T
				A  *bytes.Buffer   `name:"Annotated"`
				G  []*bytes.Buffer `group:"ALovelyGroup"`
			}) error {
				invoke3Called = true
				return nil
			},
		)

		s = Symbols{
			Names: []interface{}{
				"Constructor1",
				"Constructor2",
				Annotated{
					Name:   "Annotated",
					Target: "Target1",
				},
				Annotated{
					Group:  "ALovelyGroup",
					Target: "Target2",
				},
				"Invoke1",
				"Invoke2",
				"Invoke3",
			},
		}

		app = fxtest.New(
			suite.T(),
			s.Load(sm),
		)
	)

	app.RequireStart()

	suite.True(invoke1Called)
	suite.True(invoke2Called)
	suite.True(invoke3Called)

	app.RequireStop()
}

func (suite *SymbolsSuite) testLoadInvalidTarget() {
	testCases := []*SymbolMap{
		NewSymbols(
			"InvalidTarget", func() {},
		),
		NewSymbols(
			"InvalidTarget", func() (int, float64, error) {
				return 0, 0.0, errors.New("This shouldn't have been called")
			},
		),
		NewSymbols(
			"InvalidTarget", func() (error, error) {
				return errors.New("This shouldn't have been called"),
					errors.New("This shouldn't have been called")
			},
		),
		NewSymbols(
			"InvalidTarget", func() (int, float64) {
				return 0, 0.0
			},
		),
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			app := fx.New(
				Symbols{
					Names: []interface{}{
						Annotated{
							Name:   "Invalid",
							Target: "InvalidTarget",
						},
					},
				}.Load(testCase),
			)

			err := app.Err()
			suite.Require().Error(err)

			var ite *InvalidTargetError
			suite.Require().True(errors.As(err, &ite))
			suite.NotEmpty(ite.Error())
		})
	}
}

func (suite *SymbolsSuite) testLoadMissing() {
	suite.Run("Error", func() {
		app := fx.New(
			Symbols{
				Names: []interface{}{
					"Missing",
				},
			}.Load(NewSymbols()),
		)

		suite.Error(app.Err())
	})

	suite.Run("Ignore", func() {
		app := fxtest.New(
			suite.T(),
			Symbols{
				Names: []interface{}{
					"Missing",
				},
				IgnoreMissing: true,
			}.Load(NewSymbols()),
		)

		app.RequireStart()
		app.RequireStop()
	})
}

func (suite *SymbolsSuite) testLoadNotAFunction() {
	app := fx.New(
		Symbols{
			Names: []interface{}{
				"NotAFunction",
			},
		}.Load(NewSymbols(
			"NotAFunction", 123,
		)),
	)

	suite.Error(app.Err())
}

func (suite *SymbolsSuite) testLoadInvalidName() {
	app := fx.New(
		Symbols{
			Names: []interface{}{
				38472983712,
			},
		}.Load(NewSymbols()),
	)

	suite.Error(app.Err())
}

func (suite *SymbolsSuite) TestLoad() {
	suite.Run("Success", suite.testLoadSuccess)
	suite.Run("InvalidTarget", suite.testLoadInvalidTarget)
	suite.Run("Missing", suite.testLoadMissing)
	suite.Run("NotAFunction", suite.testLoadNotAFunction)
	suite.Run("InvalidName", suite.testLoadInvalidName)
}

func TestSymbols(t *testing.T) {
	suite.Run(t, new(SymbolsSuite))
}
