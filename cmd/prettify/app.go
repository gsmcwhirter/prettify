package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gsmcwhirter/go-util/v7/cli"
	"github.com/tidwall/gjson"
)

var levelColors = map[string]func(string, ...interface{}) string{
	"NONE": color.BlackString,
	"DEBU": color.MagentaString,
	"INFO": color.GreenString,
	"WARN": color.YellowString,
	"ERRO": color.RedString,
}

var autoFields = map[string]bool{
	"caller": true,
}

type app struct {
	cli            *cli.Command
	messageField   string
	timestampField string
	levelField     string
	stackField     string
	output         []string
	exclude        []string
	forceColor     bool
	autoFields     bool
	allStacks      bool
	skipStacks     bool
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
		"Only seeing some fields (just message in this case)", fmt.Sprintf("my-cmd | %[1]s -O 'message'", AppName),
	)

	c.SetRunFunc(a.run)
	c.Flags().StringVarP(&a.messageField, "message-field", "m", "message", "The name of a field that contains the 'message'")
	c.Flags().StringVarP(&a.timestampField, "timestamp-field", "t", "timestamp", "The name of the timestamp field")
	c.Flags().StringVarP(&a.levelField, "level-field", "l", "level", "The name of the field containing the log level")
	c.Flags().StringVarP(&a.stackField, "stack-field", "k", "stack", "The name of the field containing the stack trace")
	c.Flags().StringSliceVarP(&a.output, "output", "O", nil, "A list of fields to show (all when not present)")
	c.Flags().StringSliceVarP(&a.exclude, "exclude", "E", nil, "A list of fields to exclude (none when not present; takes priority over everything else)")
	c.Flags().BoolVarP(&a.forceColor, "color", "C", false, "Force color output (for less and similar pipes)")
	c.Flags().BoolVarP(&a.autoFields, "auto-fields", "A", false, "Include auto-generated tags from log lines (without, can still explicitly specify in -O)")
	c.Flags().BoolVarP(&a.allStacks, "all-stacks", "s", false, "Include printing a stack trace for non-error lines where it is included")
	c.Flags().BoolVarP(&a.skipStacks, "no-stacks", "S", false, "Skip printing a stack trace for lines where it is included")

	a.cli = c

	return c
}

func (a *app) run(cmd *cli.Command, args []string) error {
	if a.forceColor {
		color.NoColor = false
	}

	scanner := bufio.NewScanner(os.Stdin)

	var line string
	var obj gjson.Result
	var ts string
	var level, levelText string
	var message string

	var lineKeys []string
	var lineMap map[string]gjson.Result

	specialFields := map[string]bool{
		a.messageField:   true,
		a.timestampField: true,
		a.levelField:     true,
		a.stackField:     true,
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

		lineKeys = make([]string, 0, len(lineKeys))
		lineMap = map[string]gjson.Result{}

		obj = gjson.Parse(line)
		obj.ForEach(func(key, value gjson.Result) bool {
			kStr := key.String()
			if kStr == "" {
				return true
			}

			lineKeys = append(lineKeys, kStr)
			lineMap[kStr] = value
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
			level = strings.ToUpper(lineMap[a.levelField].String())
		} else {
			level = "NONE"
		}

		levelText = level[:4]
		if levelColor, ok := levelColors[levelText]; ok {
			levelText = levelColor("%s", levelText)
		}

		message = ""
		if !strings.HasPrefix(line, "{") {
			message = line
		} else if lineMap[a.messageField].Exists() {
			message = lineMap[a.messageField].String()
		}

		fmt.Printf("%s |%s| %s", ts, levelText, message)

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

		if lineMap[a.stackField].Exists() && !a.skipStacks && (a.allStacks || level == "ERROR") {
			rawLines := lineMap[a.stackField].Array()
			lines := make([]string, 0, len(rawLines))

			for _, l := range rawLines {
				lines = append(lines, l.String())
			}

			fmt.Printf(" %s=\n\t%s", color.CyanString(a.stackField), strings.Join(lines, "\n\t"))
		}

		fmt.Println()
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
