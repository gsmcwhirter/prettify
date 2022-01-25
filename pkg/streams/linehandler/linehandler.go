package linehandler

import (
	"fmt"
	"strings"

	"github.com/gsmcwhirter/prettify/pkg/streams/formatter"
)

// LineHandler is an interface for things that handle formatting and possibly skipping lines that should be considered for printing
type LineHandler interface {
	HandleLine(filename, line string) bool
}

// FilterLineHandler is an interface for LineHandlers that can additionally filter lines for more than being blank
type FilterLineHandler interface {
	LineHandler
}

// linePrinter is a FilterLineHandler implementation
type linePrinter struct {
	withPath     string
	withPretty   bool
	withColor    bool
	withSort     bool
	withBlanks   bool
	withFilename bool
	printf       func(string, ...interface{}) (int, error)
}

// Options controls the behavior of NewLinePrinter-created objects
//
// WithBlanks controls whether empty lines are filtered out.
// WithFilename controls whether each printed line is prefixed with the filename it came from or not.
// WithMemusage controls whether the tsar memusage line(s) will be printed
// JSONPath determines the output format (raw line if this is empty)
// Pretty determines whether the lines will attempted to be made pretty (field per line, etc)
// Color determines whether the lines are colorized or not
// Sort determines whether the keys of a json line will be sorted or not
type Options struct {
	WithBlanks   bool
	WithFilename bool
	Pretty       bool
	Color        bool
	Sort         bool
	JSONPath     string
	Printf       func(string, ...interface{}) (int, error)
}

// NewLinePrinter returns a new struct that implements the FilterLineHandler interface
func NewLinePrinter(opts Options) FilterLineHandler {
	lp := &linePrinter{
		withPath:     opts.JSONPath,
		withPretty:   opts.Pretty,
		withColor:    opts.Color,
		withSort:     opts.Sort,
		withBlanks:   opts.WithBlanks,
		withFilename: opts.WithFilename,
		printf:       opts.Printf,
	}

	if lp.printf == nil {
		lp.printf = fmt.Printf
	}

	return lp
}

// HandleLine considers printing a line and handles formatting if it will print it
func (lp *linePrinter) HandleLine(filename, line string) (lineHadNewline bool) {
	maybeNewline := ""
	l := strings.TrimRight(line, "\n")
	if l != line {
		maybeNewline = "\n"
		lineHadNewline = true
	}

	lp.maybePrint(filename, l, maybeNewline)

	return lineHadNewline
}

func (lp *linePrinter) maybePrint(filename, line, maybeNewline string) {
	var toPrint string

	if lp.withPath == "" {
		if lp.withPretty {
			toPrint = formatter.PrettyLine(line, lp.withColor, lp.withSort)
			toPrint = strings.TrimRight(toPrint, "\n")
		} else {
			toPrint = line
		}
	} else {
		toPrint = formatter.FormatLine(line, lp.withPath, lp.withPretty, lp.withColor, lp.withSort)
	}

	if !lp.withBlanks && strings.TrimSpace(toPrint) == "" {
		return
	}

	var err error
	if lp.withFilename {
		_, err = lp.printf("%s: %s%s", filename, toPrint, maybeNewline)
	} else {
		_, err = lp.printf("%s%s", toPrint, maybeNewline)
	}

	if err != nil {
		panic(err)
	}
}
