package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gsmcwhirter/go-util/v9/cli"

	"github.com/gsmcwhirter/prettify/pkg/files/finder"
	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
	"github.com/gsmcwhirter/prettify/pkg/files/streamer"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
)

type catCommand struct {
	cli         *cli.Command
	fileFinder  *finder.Finder
	linePrinter linehandler.FilterLineHandler

	JSONPath     string
	JSONPretty   bool
	JSONColor    bool
	JSONSort     bool
	WithBlanks   bool
	WithFilename bool
}

func (cmd *catCommand) catFile(ctx context.Context) func(string, os.FileInfo, error) error {
	return func(path string, _ os.FileInfo, err error) error {
		// debug print
		// fmt.Printf("Trying to cat file %s\n", path)

		if err != nil {
			return err
		}

		dirName, fileName := filepath.Split(path)
		_, err = streamer.Cat(ctx, dirName, fileName, cmd.linePrinter)

		return err
	}
}

// Run executes the `cat` command
func (cmd *catCommand) run(c *cli.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("you must provide a file pattern")
	}
	filePattern := args[0]
	if filePattern == "" {
		return errors.New("you must provide a file pattern")
	}

	ctx := context.Background()

	cmd.linePrinter = linehandler.NewLinePrinter(linehandler.Options{
		WithBlanks:   cmd.WithBlanks,
		WithFilename: cmd.WithFilename,
		JSONPath:     cmd.JSONPath,
		Pretty:       cmd.JSONPretty,
		Color:        cmd.JSONColor,
		Sort:         cmd.JSONSort,
	})

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

	return fp.WalkFiles(ctx, cmd.catFile(ctx))
}

func setupCat(c *cli.Command, appName string, fileFinder *finder.Finder) {
	c.AddExamples("Cat the contents of all matching files to stdout, skipping blank lines", fmt.Sprintf("%s cat <filepat>", appName))

	opts := &catCommand{}

	cat := cli.NewCommand("cat", cli.CommandOptions{
		ShortHelp:    "Cat the contents of some tsar-generated log files",
		PosArgsUsage: "<filepat>",
		Args:         cli.ExactArgs(1),
	})

	cat.AddExamples(
		"Cat the contents of all matching files to stdout, skipping blank lines", fmt.Sprintf("%[1]s cat <filepat>", appName),
		"Cat the contents of all matching files, skipping blank lines, prefixing each line with the filename the line came from", fmt.Sprintf("%[1]s cat <filepat> --with-filename", appName),
		"Cat the contents of all matching files to stdout, preserving blank lines", fmt.Sprintf("%[1]s cat <filepat> --with-blanks", appName),
		"Cat the contents of the matching files to stdout, selecting only the @timestamp, @tag, and message fields from each json line, and outputting the data as tab-separated values", fmt.Sprintf("%[1]s cat <filepat> --output='@timestamp,@tag,message,|@tsv'", appName),
		"Cat the contents of all matching files to stdout, skipping files that have a filename datetime after 2018010315*", fmt.Sprintf("%[1]s cat <filepat> --before='2018-01-03 15:'", appName),
		"Cat the contents of all matching files to stdout, skipping files that have a filename datetime before 2018010315*", fmt.Sprintf("%[1]s cat <filepat> --since='2018-01-03 15:'", appName),
	)

	cat.SetRunFunc(opts.run)

	cat.Flags().StringVarP(&opts.JSONPath, "output", "O", "", "An output expression (selects which fields to show and how)")
	cat.Flags().BoolVarP(&opts.JSONPretty, "pretty", "P", false, "Pretty-print json lines")
	cat.Flags().BoolVarP(&opts.JSONColor, "color", "C", false, "Add color to pretty-printed json lines")
	cat.Flags().BoolVarP(&opts.JSONSort, "sort", "S", false, "Sort output keys")
	cat.Flags().BoolVar(&opts.WithBlanks, "with-blanks", false, "Include blank lines in the output")
	cat.Flags().BoolVar(&opts.WithFilename, "with-filename", false, "Display the filename at the beginning of each line")

	c.AddSubCommands(cat)

	opts.cli = cat
	opts.fileFinder = fileFinder
}
