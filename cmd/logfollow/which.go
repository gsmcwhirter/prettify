package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gsmcwhirter/go-util/v9/cli"

	"github.com/gsmcwhirter/prettify/pkg/files/finder"
	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
)

type whichCommand struct {
	cli        *cli.Command
	fileFinder *finder.Finder

	FindAll bool
}

func (cmd *whichCommand) findOne(ctx context.Context, filePattern string) (bool, error) {
	fp, err := cmd.fileFinder.Find(ctx, filePattern)
	if err != nil {
		return false, err
	}

	if fp != nil {
		fmt.Println(fp.Which())
		return true, nil
	}

	return false, nil
}

func (cmd *whichCommand) findAll(ctx context.Context, filePattern string) (bool, error) {
	fpList, err := cmd.fileFinder.FindAllFromPattern(ctx, filePattern)
	if err != nil {
		return false, err
	}

	for _, fp := range fpList {
		if fp != nil {
			fmt.Println(fp.Which())
		}
	}

	return true, nil
}

func (cmd *whichCommand) run(c *cli.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("you must provide a file pattern")
	}
	filePattern := args[0]
	if filePattern == "" {
		return errors.New("you must provide a file pattern")
	}

	ctx := context.Background()

	var found bool
	var err error
	if cmd.fileFinder != nil {
		if cmd.FindAll {
			found, err = cmd.findAll(ctx, filePattern)
		} else {
			found, err = cmd.findOne(ctx, filePattern)
		}

		if err != nil {
			return err
		}
	}

	if !found {
		fp := pattern.NewPattern(filePattern)
		fmt.Println(fp.Which())
	}

	return nil
}

// Setup is responsible for initializing the `cat` command options
func setupWhich(c *cli.Command, appName string, fileFinder *finder.Finder) {
	c.AddExamples("Display a glob representing which files would be processed", fmt.Sprintf("%s which <filepat>", appName))

	opts := &whichCommand{}
	which := cli.NewCommand("which", cli.CommandOptions{
		ShortHelp:    "Display a glob representing which files would be processed",
		PosArgsUsage: "<filepat>",
		Args:         cli.ExactArgs(1),
	})

	which.AddExamples(
		"Display a glob representing which files would be processed", fmt.Sprintf("%s which <filepat>", appName),
		"Display a glob representing all files that you might have meant", fmt.Sprintf("%s which <filepat> --all", appName),
	)

	which.SetRunFunc(opts.run)

	which.Flags().BoolVarP(&opts.FindAll, "all", "a", false, "Display _all_ globs that match the pattern")

	c.AddSubCommands(which)

	opts.cli = which
	opts.fileFinder = fileFinder
}
