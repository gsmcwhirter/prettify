package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gsmcwhirter/go-util/v9/cli"

	"github.com/gsmcwhirter/prettify/pkg/files/finder"
)

func optionParser() *cli.Command {
	ff := finder.NewFinder([]string{
		".",
		path.Join(os.Getenv("HOME"), ".pm2/logs"),
	})

	c := cli.NewCLI(AppName, BuildVersion, BuildSHA, BuildDate, cli.CommandOptions{
		ShortHelp: "Gather information about tsar-generated logs",
		LongHelp: fmt.Sprintf(`Gather information about logs.

  Searches the following directories for <filepat> arguments:
    - %s

  Some commands take output arguments.

  Output Expression: <field name>[,<field name>...][,<formatter>]
    - Field names that are specified but not present in a line will be treated as an empty string
    - You might need to escape @ symbols in field names depending on your shell
    - Valid <formatter> expressions include (with leading '|' character):
	  |@tsv (tab-separated values)
	  |@csv (comma-separated values)
	  |@ssv (space-separated values)
	  |@nlsv (newline-separated values; default)

`, strings.Join(ff.SearchDirectories, "\n    - ")),
		Example: "",
	})

	setupCat(c, AppName, ff)
	setupTac(c, AppName, ff)
	setupTail(c, AppName, ff)
	setupWhich(c, AppName, ff)
	setupFind(c, AppName, ff)

	return c
}
