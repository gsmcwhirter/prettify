package main

import (
	"fmt"
	"os"
)

// Build-time variables
var (
	AppName      string
	BuildDate    string
	BuildVersion string
	BuildSHA     string
)

func main() {
	code, err := run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", AppName, err)
	}

	os.Exit(code)
}

// Separate this function such that defers, which are skipped on os.Exit(), are run
func run() (int, error) {
	code := 0

	a := app{}
	cli := a.setup()

	err := cli.Execute()
	if err != nil {
		return -1, err
	}

	return code, nil
}
