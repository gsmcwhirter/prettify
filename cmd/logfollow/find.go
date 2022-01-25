package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gsmcwhirter/go-util/v9/cli"

	"github.com/gsmcwhirter/prettify/internal/tagfinder"
	"github.com/gsmcwhirter/prettify/pkg/files/finder"
	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
)

type findCommand struct {
	cli        *cli.Command
	fileFinder *finder.Finder

	FilePattern string
	Tag         string
	FindAll     bool
	SampleSize  uint
}

func (cmd *findCommand) printPatternTags(ctx context.Context, fp *pattern.Pattern) {
	tags := make([]string, 0)

	if fp != nil {
		fpTags, err := cmd.findTags(ctx, fp)
		if err != nil {
			fmt.Println(err)
			return
		}

		tags = append(tags, fpTags...)
	}

	fmt.Printf("@tag values in %s:\n", fp.Which())
	for _, tag := range tags {
		fmt.Printf("  %s\n", tag)
	}
}

func tagInList(targetTag string, tags []string) bool {
	for _, tag := range tags {
		if tag == targetTag {
			return true
		}
	}

	return false
}

func (cmd *findCommand) printTagFiles(ctx context.Context, targetTag string) error {
	fpList, err := cmd.fileFinder.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, fp := range fpList {
		tags, err := cmd.findTags(ctx, fp)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if tagInList(targetTag, tags) {
			fmt.Println(fp.Pattern())
		}
	}

	return nil
}

func (cmd *findCommand) findTags(ctx context.Context, fp *pattern.Pattern) ([]string, error) {
	tf := tagfinder.NewTagFinder(cmd.SampleSize, cmd.FindAll)
	err := fp.WalkFiles(ctx, tf.Walker)
	if err != nil {
		return nil, nil
	}

	return tf.Tags(), nil
}

func (cmd *findCommand) printFileTags(ctx context.Context, filePattern string) error {
	fpList, err := cmd.fileFinder.FindAllFromPattern(ctx, filePattern)
	if err != nil {
		return err
	}

	for _, fp := range fpList {
		cmd.printPatternTags(ctx, fp)
		fmt.Println()
	}

	return nil
}

func (cmd *findCommand) printFilesAndTags(ctx context.Context) error {
	fpList, err := cmd.fileFinder.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, fp := range fpList {
		cmd.printPatternTags(ctx, fp)
		fmt.Println()
	}

	return nil
}

// Run executes the `find` command
func (cmd *findCommand) run(c *cli.Command, args []string) error {
	ctx := context.Background()

	if cmd.FilePattern != "" && cmd.Tag != "" {
		return errors.New("you should not provide both a file pattern and a tag")
	}

	if cmd.FilePattern != "" {
		return cmd.printFileTags(ctx, cmd.FilePattern)
	}

	if cmd.Tag != "" {
		return cmd.printTagFiles(ctx, cmd.Tag)
	}

	return cmd.printFilesAndTags(ctx)
}

// Setup is responsible for initializing the `cat` command options
func setupFind(c *cli.Command, appName string, fileFinder *finder.Finder) {
	c.AddExamples("Display file patterns and tags that tlogs could find", fmt.Sprintf("%s find", appName))

	opts := &findCommand{}

	find := cli.NewCommand("find", cli.CommandOptions{
		ShortHelp: "Display file patterns and tags that tlogs could find",
		Args:      cli.NoArgs,
	})

	find.AddExamples(
		"Display file patterns and tags that tlogs could find", fmt.Sprintf("%[1]s find", appName),
		"Display a sample list of tags that appear in files of the specified file pattern", fmt.Sprintf("%[1]s find --filepat=<filepat>", appName),
		"Display a comprehensive list of tags that appear in files of the specified file pattern", fmt.Sprintf("%[1]s find --filepat=<filepat> -a", appName),
		"Display a sample list of files in which the given tag appears", fmt.Sprintf("%[1]s find --tag=app.search-team.lager", appName),
		"Display a comprehensive list of files in which the given tag appears", fmt.Sprintf("%[1]s find --tag=app.search-team.lager -a", appName),
	)

	find.SetRunFunc(opts.run)

	find.Flags().StringVarP(&opts.FilePattern, "filepat", "f", "", "Find tags in the given file pattern")
	find.Flags().StringVarP(&opts.Tag, "tag", "t", "", "Find files containing the given tag")
	find.Flags().BoolVarP(&opts.FindAll, "all", "a", false, "Display _all_ results (don't sample)")
	find.Flags().UintVarP(&opts.SampleSize, "samples", "n", 1000, "Number of samples to take from files without --all")

	c.AddSubCommands(find)

	opts.cli = find
	opts.fileFinder = fileFinder
}
