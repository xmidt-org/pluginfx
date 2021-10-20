package pluginfx

import (
	"fmt"
	"os"
	"os/exec"
	"plugin"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const samplePath = "sample.so"

func TestMain(m *testing.M) {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "./sample")
	fmt.Println(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to build sample plugin: %s\n", err)
		os.Exit(1)
	}

	var code int
	defer func() {
		os.Remove(samplePath)
		os.Exit(code)
	}()

	code = m.Run()
}

func TestSample(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		p, err  = plugin.Open(samplePath)
	)

	require.NoError(err)
	require.NotNil(p)

	globalInt, err := p.Lookup("Value")
	require.NoError(err)
	assert.Equal(12, *globalInt.(*int))

	f, err := p.Lookup("New")
	require.NoError(err)

	ft := reflect.TypeOf(f)
	require.Equal(reflect.Func, ft.Kind())

	floatValue, err := f.(func() (float64, error))()
	require.NoError(err)
	assert.Equal(67.5, floatValue)
}
