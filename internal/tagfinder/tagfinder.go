package tagfinder

import (
	"bufio"
	"errors"
	"os"
	"sort"

	"github.com/tidwall/gjson"

	"github.com/gsmcwhirter/go-util/v9/deferutil"
)

func init() {
	gjson.DisableModifiers = true
}

// TagFinder handles reading files as they are iterated through by a directory walker
// and finding the tags that exist
type TagFinder struct {
	tags       map[string]bool
	sampleSize uint
	findAll    bool
}

// NewTagFinder creates a new TagFinder
func NewTagFinder(sampleSize uint, findAll bool) TagFinder {
	return TagFinder{
		tags:       map[string]bool{},
		sampleSize: sampleSize,
		findAll:    findAll,
	}
}

// Tags returns a sorted list of tags that have been seen
func (tf *TagFinder) Tags() []string {
	tags := make([]string, len(tf.tags))

	i := 0
	for key := range tf.tags {
		tags[i] = key
		i++
	}

	sort.Strings(tags)
	return tags
}

func extractTag(line string) (string, error) {
	res := gjson.Get(line, "@tag")
	if !res.Exists() {
		return "", errors.New("no @tag found")
	} else if res.Type != gjson.String {
		return "", errors.New("@tag not string")
	}

	return res.String(), nil
}

// Walker handles reading files and finding tags as they are iterated through by a directory walker
func (tf *TagFinder) Walker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer deferutil.CheckDefer(file.Close)

	scanner := bufio.NewScanner(file)
	var lineCt uint
	for scanner.Scan() {
		line := scanner.Text()
		lineCt++

		tag, err := extractTag(line)
		if err != nil {
			continue
		}

		tf.tags[tag] = true

		if !tf.findAll && lineCt >= tf.sampleSize {
			break
		}
	}

	return nil
}
