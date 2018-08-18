package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"strings"

	"github.com/gsmcwhirter/go-util/deferutil"
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

	// Initiate a runtime trace
	tracePath := getTracePath()
	if tracePath != "" {
		tf, err := os.Create(tracePath)
		if err != nil {
			return -1, err
		}
		defer deferutil.CheckDefer(tf.Close)

		err = trace.Start(tf)
		if err != nil {
			return -1, err
		}
		defer trace.Stop()
	}

	a := app{}
	cli := a.setup()

	err := cli.Execute()
	if err != nil {
		return -1, err
	}

	return code, nil
}

func getTracePath() string {
	// Hidden option to produce a trace of the runtime
	if val, ok := os.LookupEnv(fmt.Sprintf("%s_TRACE", strings.ToUpper(AppName))); ok && val != "" {
		if val == "." {
			return fmt.Sprintf("%s.trace", AppName)
		}

		return val
	}

	return ""
}
