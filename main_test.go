package pluginfx

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
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
