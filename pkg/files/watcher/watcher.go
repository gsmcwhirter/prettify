// Package watcher contains functionality for watching
// for new files matching a files/pattern FilePattern
// and applying actions to the new files
package watcher

import (
	"context"
	"errors"
	"os"

	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
)

// Watcher watches for new files matching a files/pattern FilePattern
//
// Note: this has unbounded memory in theory, though that memory should grow
// rather slowly unless there are lots of new files being created that match
// a FilePattern
type Watcher struct {
	SeenFiles      map[string]bool
	seenThisTime   []string
	seenThisTimeCt int
	filePattern    *pattern.Pattern
	lastFile       string
	prevLastFile   string
}

// NewWatcher creates a new watcher from a FilePattern
func NewWatcher(fp *pattern.Pattern) *Watcher {
	fw := Watcher{
		SeenFiles:      map[string]bool{},
		seenThisTime:   make([]string, 16),
		seenThisTimeCt: 0,
		filePattern:    fp,
	}

	return &fw
}

func (fw *Watcher) resetSeenThisTime() {
	fw.seenThisTimeCt = 0
}

func (fw *Watcher) seeFile(path string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	_, present := fw.SeenFiles[path]
	if !present {
		fw.addToSeenThisTime(path)
		fw.lastFile = path
	}

	fw.SeenFiles[path] = true

	return nil
}

func (fw *Watcher) addToSeenThisTime(path string) {
	for fw.seenThisTimeCt >= len(fw.seenThisTime) {
		newSeenThisTime := make([]string, 2*len(fw.seenThisTime))
		for i := 0; i < fw.seenThisTimeCt; i++ {
			newSeenThisTime[i] = fw.seenThisTime[i]
		}
		fw.seenThisTime = newSeenThisTime
	}

	fw.seenThisTime[fw.seenThisTimeCt] = path
	fw.seenThisTimeCt++
}

// LastSeenThisTime gets the last filename that was seen on the current run
func (fw *Watcher) LastSeenThisTime() (string, error) {
	if fw.seenThisTimeCt == 0 {
		return "", errors.New("no new files seen this time")
	}

	return fw.seenThisTime[fw.seenThisTimeCt-1], nil
}

// SeenThisTime returns the list of not-seen-before filenames from this run
func (fw *Watcher) SeenThisTime() []string {
	ret := []string{}

	if fw.prevLastFile != "" {
		ret = append(ret, fw.prevLastFile)
	}

	if fw.seenThisTimeCt == 0 {
		return ret
	}

	return append(ret, fw.seenThisTime[0:fw.seenThisTimeCt]...)
}

// Run kicks off a round of watching the directory
func (fw *Watcher) Run(ctx context.Context) error {
	fw.resetSeenThisTime()
	fw.prevLastFile = fw.lastFile
	err := fw.filePattern.WalkFiles(ctx, fw.seeFile)
	return err
}
