package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gsmcwhirter/go-util/cli"
	"github.com/tidwall/gjson"

	"github.com/gsmcwhirter/prettify/pkg/filter"
)

var levelColors = map[string]func(string, ...interface{}) string{
	"NONE": color.BlackString,
	"DEBU": color.MagentaString,
	"INFO": color.GreenString,
	"WARN": color.YellowString,
	"ERRO": color.RedString,
}

var autoFields = map[string]bool{
	"@tag":            true,
	"PID":             false, // yes, false -- we usually want this
	"event_timestamp": true,
	"file_name":       true,
	"function_name":   true,
	"line_number":     true,
	"log_level":       true,
	"main_name":       true,
	"script_exec_id":  true,
	"script_exec_nth": true,
	"source_host":     true,
	"time_iso":        true,
}

type app struct {
	cli            *cli.Command
	messageFields  []string
	timestampField string
	levelField     string
	filters        []string
	output         []string
	exclude        []string
	forceColor     bool
	autoFields     bool
}

func (a *app) setup() *cli.Command {
	// TODO: example(s)

	c := cli.NewCLI(AppName, BuildVersion, BuildSHA, BuildDate, cli.CommandOptions{
		ShortHelp: `Transform json log lines into a prettier format`,
		LongHelp: `Transform json log lines into a prettier format

  This accepts input on stdin and writes back to stdout.
  You might want to use a 2>&1 construct to pipe stdout and stderr through the same invocation.`,
		Args: cli.NoArgs,
	})

	c.AddExamples(
		"Just seeing logs", fmt.Sprintf("my-cmd 2>&1 | %[1]s", AppName),
		"Saving raw logs and seeing nice versions", fmt.Sprintf("my-cmd 2>&1 | tee -a real_data.log | %[1]s", AppName),
		"Filtering logs", fmt.Sprintf("my-cmd | %[1]s -F 'log_level > 20'", AppName),
		"Only seeing some fields (just message in this case)", fmt.Sprintf("my-cmd | %[1]s -O 'message'", AppName),
	)

	c.SetRunFunc(a.run)
	c.Flags().StringSliceVarP(&a.messageFields, "message-field", "m", []string{"message", "msg"}, "The name of a field that contains the 'message'")
	c.Flags().StringVarP(&a.timestampField, "timestamp-field", "t", "@timestamp", "The name of the timestamp field")
	c.Flags().StringVarP(&a.levelField, "level-field", "l", "log_level_string", "The name of the field containing the log level")
	c.Flags().StringArrayVarP(&a.filters, "filter", "F", []string{}, "A filter expression (lines not matching will not be displayed -- repeatable)")
	c.Flags().StringSliceVarP(&a.output, "output", "O", nil, "A list of fields to show (all when not present)")
	c.Flags().StringSliceVarP(&a.exclude, "exclude", "E", nil, "A list of fields to exclude (none when not present; takes priority over everything else)")
	c.Flags().BoolVarP(&a.forceColor, "color", "C", false, "Force color output (for less and similar pipes)")
	c.Flags().BoolVarP(&a.autoFields, "auto-fields", "A", false, "Include auto-generated tags from log lines (without, can still explicitly specify in -O)")

	a.cli = c

	return c
}

func (a *app) run(cmd *cli.Command, args []string) error {
	if a.forceColor {
		color.NoColor = false
	}

	lf := filter.NewLineFilter(a.filters)

	scanner := bufio.NewScanner(os.Stdin)

	var line string
	var obj gjson.Result
	var ts string
	var level string
	var message string

	var lineKeys []string
	var lineMap map[string]gjson.Result

	specialFields := map[string]bool{
		a.timestampField: true,
		a.levelField:     true,
	}
	for _, mField := range a.messageFields {
		specialFields[mField] = true
	}

	outputFields := map[string]bool{}
	for _, oField := range a.output {
		outputFields[oField] = true
	}

	excludeFields := map[string]bool{}
	for _, eField := range a.exclude {
		excludeFields[eField] = true
	}

	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())

		if !lf.Allow(line) {
			continue
		}

		lineKeys = make([]string, 0, len(lineKeys))
		lineMap = map[string]gjson.Result{}

		obj = gjson.Parse(line)
		obj.ForEach(func(key, value gjson.Result) bool {
			lineKeys = append(lineKeys, key.String())
			lineMap[key.String()] = value
			return true // keep iterating
		})

		sort.Strings(lineKeys)

		if lineMap[a.timestampField].Exists() {
			ts = lineMap[a.timestampField].String()
		} else {
			ts = ""
		}

		ts = color.BlueString(ts)

		if lineMap[a.levelField].Exists() {
			level = lineMap[a.levelField].String()
		} else {
			level = "NONE"
		}

		level = level[:4]
		if levelColor, ok := levelColors[level]; ok {
			level = levelColor("%s", level)
		}

		message = ""
		if !strings.HasPrefix(line, "{") {
			message = line
		} else {
			for _, k := range a.messageFields {
				if lineMap[k].Exists() {
					message = lineMap[k].String()
					break
				}
			}
		}

		fmt.Printf("%s |%s| %s", ts, level, message)

		for _, key := range lineKeys {
			if _, ok := specialFields[key]; ok {
				continue
			}

			if excludeFields[key] {
				continue
			}

			if len(a.output) > 0 && !outputFields[key] { // only display requested fields
				continue
			}

			if !a.autoFields && autoFields[key] && (len(a.output) == 0 || outputFields[key]) { // get rid of any non-requested auto-fields, unless --auto-fields
				continue
			}

			fmt.Printf(" %s=%s", color.CyanString(key), lineMap[key].String())
		}

		fmt.Println()
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
