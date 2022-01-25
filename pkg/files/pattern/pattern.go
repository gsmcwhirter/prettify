package pattern

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Pattern holds information relevant to picking out certain tsar logs
type Pattern struct {
	origPattern   string
	directory     string
	filePattern   string
	filenameGlob  string
	filenameRegex *regexp.Regexp
}

// NewPattern constructs a Pattern
//
// pat is used to construct the filename trunk. It can contain a directory as
// part of the value or simply be a file basename.
func NewPattern(pat string) Pattern {
	dirName, fileName := filepath.Split(pat)
	requiredExtension := ""

	absDirName, err := filepath.Abs(dirName)
	if err != nil {
		panic(err) // TODO: Can panic on os.Getwd() failing
	}

	if !strings.HasSuffix(absDirName, "/") {
		absDirName += "/"
	}

	firstDotIdx := strings.IndexRune(fileName, '.')
	if firstDotIdx != -1 {
		requiredExtension = fileName[firstDotIdx+1:]
		fileName = fileName[:firstDotIdx]
	}
	fileName = strings.TrimRight(fileName, "-")

	fp := Pattern{
		origPattern:   pat,
		directory:     absDirName,
		filePattern:   fileName,
		filenameGlob:  formFileNameGlob(fileName, requiredExtension),
		filenameRegex: formFileNameRegexp(fileName, requiredExtension),
	}

	return fp
}

// NewPatternPtr constructs a Pattern via NewPattern and returns a pointer to it
//
// pat is used to construct the filename trunk. It can contain a directory as
// part of the value or simply be a file basename.
func NewPatternPtr(pat string) *Pattern {
	fp := NewPattern(pat)
	return &fp
}

// NewPatternPtrWithOriginal constructs a Pattern via NewPattern and returns a pointer to it
// after setting the origPattern field to a manual value
//
// This is mostly used for testing
//
// pat is used to construct the filename trunk. It can contain a directory as
// part of the value or simply be a file basename.
func NewPatternPtrWithOriginal(pat, origPattern string) *Pattern {
	p := NewPatternPtr(pat)
	p.origPattern = origPattern
	return p
}

// Pattern returns the pattern string that can generate this Pattern
func (fp *Pattern) Pattern() string {
	return fp.origPattern
}

// Which returns a string representing the file glob that this Pattern will find
func (fp *Pattern) Which() string {
	return fmt.Sprintf("%s%s", fp.directory, fp.filenameGlob)
}

// MatchesFile determines if the provided filename matches the pattern
func (fp *Pattern) matchesFile(dirName, fileName string) bool {
	if !strings.HasPrefix(dirName, fp.directory) {
		return false
	}

	if !fp.filenameRegex.MatchString(fileName) {
		return false
	}

	return true
}

// WalkFiles walks the Pattern directory and calls walkFunc on all matching files
func (fp *Pattern) WalkFiles(ctx context.Context, walkFunc filepath.WalkFunc) error {
	fpw := walker{
		filePattern: fp,
		walkFunc:    walkFunc,
	}

	return fpw.walk(ctx)
}

// WalkFilesReverse walks the Pattern directory in reverse and calls walkFunc on all matching files
func (fp *Pattern) WalkFilesReverse(ctx context.Context, walkFunc filepath.WalkFunc) error {
	fpw := walker{
		filePattern: fp,
		walkFunc:    walkFunc,
	}

	return fpw.walkReverse(ctx)
}
