package finder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
)

// Finder is responsible for finding where a set of tsar logs is
//
// The SearchDirectories list is searched in order
type Finder struct {
	SearchDirectories []string
	foundFiles        bool
}

// NewFinder creates a new Finder
//
// The provided searchDirs will be normalized and
// checked for existence before being added to the object
func NewFinder(searchDirs []string) *Finder {
	ff := Finder{
		SearchDirectories: make([]string, 0),
	}

	for _, dir := range searchDirs {
		ff.AppendSearchDirectory(dir)
	}

	return &ff
}

func (ff *Finder) markSeen(_ string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	ff.foundFiles = true
	return nil
}

// PrependSearchDirectory adds a search directory to the front of the list
// after cleaning and checking to make sure it exists.
func (ff *Finder) PrependSearchDirectory(dir string) {
	dir = cleanDirName(dir)
	if dirExists(dir) {
		ff.SearchDirectories = append([]string{dir}, ff.SearchDirectories...)
	}
}

// AppendSearchDirectory adds a search directory to the end of the list
// after cleaning and checking to make sure it exists.
func (ff *Finder) AppendSearchDirectory(dir string) {
	dir = cleanDirName(dir)
	if dirExists(dir) {
		ff.SearchDirectories = append(ff.SearchDirectories, dir)
	}
}

// Find attempts to find a directory where files of the requested type live
//
// The SearchDirectories are iterated in order and the first matching directory
// is used to construct the result.
func (ff *Finder) Find(ctx context.Context, filePat string) (*pattern.Pattern, error) {
	dirName, fileName := filepath.Split(filePat)

	if dirName != "" {
		fp := pattern.NewPattern(filePat)
		return &fp, nil
	}

	var err error
	for _, dirName := range ff.SearchDirectories {
		fp := pattern.NewPattern(dirName + fileName)
		if err1 := fp.WalkFiles(ctx, ff.markSeen); err1 != nil {
			err = err1
		} else if ff.foundFiles {
			return &fp, nil
		}
	}

	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("could not find the requested file pattern %s", fileName)
}

// FindAllFromPattern attempts to find _all_ directories where files of the requested type live
//
// The SearchDirectories are iterated in order and results are returned in that same order.
func (ff *Finder) FindAllFromPattern(ctx context.Context, filePat string) ([]*pattern.Pattern, error) {
	filePatterns := make([]*pattern.Pattern, 0)

	dirName, fileName := filepath.Split(filePat)

	if dirName != "" {
		fp := pattern.NewPattern(filePat)
		filePatterns = append(filePatterns, &fp)
		return filePatterns, nil
	}

	var err error
	for _, dirName := range ff.SearchDirectories {
		ff.foundFiles = false
		fp := pattern.NewPattern(dirName + fileName)
		if err1 := fp.WalkFiles(ctx, ff.markSeen); err1 != nil {
			err = err1
		} else if ff.foundFiles {
			filePatterns = append(filePatterns, &fp)
		}
	}

	return filePatterns, err
}

// FindAll attempts to detect all file patterns that one might want to use
//
// The SearchDirectories are iterated in order and results are returned in that same order.
// A result is added for each file base that is detected in any SearchDirectories entry
// that matches a tsar-generated file pattern (with more leeway on format). See PatternFinder
// for pattern details.
func (ff *Finder) FindAll(ctx context.Context) ([]*pattern.Pattern, error) {
	filePatterns := make([]*pattern.Pattern, 0)

	var err error
	for _, dirName := range ff.SearchDirectories {
		pfinder := pattern.NewFinder(dirName)

		if err1 := pfinder.Walk(ctx); err1 != nil {
			fmt.Println(err1)
			err = err1
			continue
		}

		for _, patternStr := range pfinder.SeenPatterns() {
			fp := pattern.NewPattern(patternStr)
			filePatterns = append(filePatterns, &fp)
		}
	}

	return filePatterns, err
}
