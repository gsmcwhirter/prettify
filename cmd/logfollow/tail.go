package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gsmcwhirter/go-util/v9/cli"

	"github.com/gsmcwhirter/prettify/pkg/files/finder"
	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
	"github.com/gsmcwhirter/prettify/pkg/files/streamer"
	"github.com/gsmcwhirter/prettify/pkg/files/watcher"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
)

type tailCommand struct {
	cli         *cli.Command
	fileFinder  *finder.Finder
	fileWatcher *watcher.Watcher
	linePrinter linehandler.FilterLineHandler

	JSONPath     string
	JSONPretty   bool
	JSONColor    bool
	JSONSort     bool
	Follow       bool
	WithBlanks   bool
	WithFilename bool
	NumLines     uint
}

func (cmd *tailCommand) printTail(ctx context.Context) (lastFile string, lastFilePos int64, err error) {
	// make sure cmd.fileWatcher.Run() is called first

	lastFile, err = cmd.fileWatcher.LastSeenThisTime()
	if err != nil {
		return lastFile, lastFilePos, err
	}

	lastFilePos, err = streamer.TailFiles(ctx, cmd.fileWatcher.SeenThisTime(), -int(cmd.NumLines), cmd.linePrinter)

	return lastFile, lastFilePos, err
}

func (cmd *tailCommand) run(c *cli.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("you must provide a file pattern")
	}
	filePattern := args[0]
	if filePattern == "" {
		return errors.New("you must provide a file pattern")
	}

	ctx := context.Background()

	var fp *pattern.Pattern

	if cmd.fileFinder != nil {
		fpTmp, err := cmd.fileFinder.Find(ctx, filePattern)
		if err != nil {
			return err
		}

		if fpTmp != nil {
			fp = fpTmp
		}
	}

	if fp == nil {
		fpTmp := pattern.NewPattern(filePattern)
		fp = &fpTmp
	}

	cmd.fileWatcher = watcher.NewWatcher(fp)
	cmd.linePrinter = linehandler.NewLinePrinter(linehandler.Options{
		WithBlanks:   cmd.WithBlanks,
		WithFilename: cmd.WithFilename,
		JSONPath:     cmd.JSONPath,
		Pretty:       cmd.JSONPretty,
		Color:        cmd.JSONColor,
		Sort:         cmd.JSONSort,
	})

	// Sets the SeenFiles
	err := cmd.fileWatcher.Run(ctx)
	if err != nil {
		return err
	}

	lastFile, lastFilePos, err := cmd.printTail(ctx)
	if err != nil {
		return err
	}

	if !cmd.Follow {
		return nil
	}

	tailFollower := streamer.TailFollower{
		FileWatcher: cmd.fileWatcher,
		LineHandler: cmd.linePrinter,
	}

	return tailFollower.FollowTail(ctx, lastFile, lastFilePos)
}

func setupTail(c *cli.Command, appName string, fileFinder *finder.Finder) {
	c.AddExamples("Tail the contents of matching files, showing the last 10 lines and following changes (only the newest file will be actually tailed)", fmt.Sprintf("%s tail -n 10 -f <filepat>", appName))

	opts := &tailCommand{}

	tail := cli.NewCommand("tail", cli.CommandOptions{
		ShortHelp:    "Tail the contents of some tsar-generated log files",
		PosArgsUsage: "<filepat>",
		Args:         cli.ExactArgs(1),
	})

	tail.AddExamples(
		"Tail the contents of matching files (5 lines by default)", fmt.Sprintf("%s tail <filepat>", appName),
		"Tail the contents of matching files, following new changes (only the newest file will be actually tailed)", fmt.Sprintf("%s tail -f <filepat>", appName),
		"Tail the contents of matching files, showing the last 10 lines", fmt.Sprintf("%s tail -n 10 <filepat>", appName),
		"Tail the contents of matching files, skipping blank lines, prefixing each line with the filename the line came from.", fmt.Sprintf("%s tail -n 10 <filepat> --with-filename", appName),
		"Tail the contents of matching files to stdout, preserving blank lines", fmt.Sprintf("%s tail -n 10<filepat> --with-blanks", appName),
		"Tail the contents of the matching files to stdout, selecting only the @timestamp, @tag, and message fields from each json line, and outputting the data as tab-separated values", fmt.Sprintf("%s tail -n 10 <filepat> --jj='@timestamp,@tag,message,|@tsv'", appName),
	)

	tail.SetRunFunc(opts.run)

	tail.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Follow the files")
	tail.Flags().UintVarP(&opts.NumLines, "num-lines", "n", 5, "Tail starting this many lines back")
	tail.Flags().StringVarP(&opts.JSONPath, "output", "O", "", "An output expression (selects which fields to show and how)")
	tail.Flags().BoolVarP(&opts.JSONPretty, "pretty", "P", false, "Pretty-print json lines")
	tail.Flags().BoolVarP(&opts.JSONColor, "color", "C", false, "Add color to pretty-printed json lines")
	tail.Flags().BoolVarP(&opts.JSONSort, "sort", "S", false, "Sort output keys")
	tail.Flags().BoolVar(&opts.WithBlanks, "with-blanks", false, "Include blank lines in the output")
	tail.Flags().BoolVar(&opts.WithFilename, "with-filename", false, "Display the filename at the beginning of each line")

	c.AddSubCommands(tail)

	opts.cli = tail
	opts.fileFinder = fileFinder
}
