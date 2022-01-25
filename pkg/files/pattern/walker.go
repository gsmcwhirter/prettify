package pattern

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// walker walks the filesystem looking for files that match a pattern
type walker struct {
	filePattern *Pattern
	walkFunc    filepath.WalkFunc
}

func (fpw *walker) walkFilter(ctx context.Context) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err // not sure what to do about permissions issues -- right now, might require sudo
		}

		dirName, fileName := filepath.Split(path)

		// debug print
		// fmt.Printf("Considering %s%s\n", dirName, fileName)

		if !strings.HasPrefix(fpw.filePattern.directory, dirName) {
			// debug print
			// fmt.Printf("Wrong prefix (not %s)\n", fpw.Pattern.directory)

			return filepath.SkipDir
		}

		if !fpw.filePattern.matchesFile(dirName, fileName) {
			// debug print
			// fmt.Printf("No pattern match (%v)\n", fpw.Pattern)

			return nil
		}

		// debug print
		// fmt.Println("Pattern matched.")

		return fpw.walkFunc(path, info, err)
	}
}

func (fpw *walker) walk(ctx context.Context) error {
	// debug print
	// fmt.Printf("Walking %s\n", fpw.Pattern.directory)

	return filepath.Walk(fpw.filePattern.directory, fpw.walkFilter(ctx))
}

func (fpw *walker) setWalkFunc(walkFunc filepath.WalkFunc) {
	fpw.walkFunc = walkFunc
}

func (fpw *walker) walkReverse(ctx context.Context) error {
	revAcc := reverseAccumulator{
		Paths: []reverseAccumulatorRecord{},
	}

	origwalker := fpw.walkFunc
	fpw.setWalkFunc(revAcc.record)

	err := filepath.Walk(fpw.filePattern.directory, fpw.walkFilter(ctx))
	if err != nil {
		return err
	}

	fpw.walkFunc = origwalker

	sort.Sort(&revAcc)

	for _, rar := range revAcc.Paths {
		err := fpw.walkFunc(rar.path, rar.info, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
