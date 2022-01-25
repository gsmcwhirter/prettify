package pattern

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var patternFilenameRegex = regexp.MustCompile(`^([^.]+)-(?:out|error)(?:-\d+)?(.*)$`)

// Finder will scan a filesystem directory for files that a FileFinder might find
type Finder struct {
	directory    string
	seenPatterns map[string]bool
	matchRegex   *regexp.Regexp
}

// NewFinder creates a new Finder
func NewFinder(directory string) Finder {
	return Finder{
		directory:    directory,
		seenPatterns: map[string]bool{},
		matchRegex:   patternFilenameRegex,
	}
}

// SeenPatterns returns a sorted list of pattern strings that were found
func (pf *Finder) SeenPatterns() []string {
	patterns := make([]string, len(pf.seenPatterns))

	i := 0
	for pattern := range pf.seenPatterns {
		patterns[i] = pattern
		i++
	}

	sort.Strings(patterns)
	return patterns
}

func (pf *Finder) walkFilter(ctx context.Context) func(string, os.FileInfo, error) error {
	return func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if pf.matchRegex == nil {
			return errors.New("could not patternWalker.walkFilter without a matchRegex")
		}

		dirName, fileName := filepath.Split(path)

		if !strings.HasPrefix(pf.directory, dirName) {
			return filepath.SkipDir
		}

		if pf.matchRegex.MatchString(fileName) {
			matches := pf.matchRegex.FindStringSubmatch(fileName)

			filePat := matches[1]
			fileExt := matches[2]

			pf.seenPatterns[fmt.Sprintf("%s%s%s", dirName, filePat, fileExt)] = true
		}

		return nil
	}
}

// Walk will fill the SeenFiles array by walking the filesystem in the provided directory
func (pf *Finder) Walk(ctx context.Context) error {
	return filepath.Walk(pf.directory, pf.walkFilter(ctx))
}
