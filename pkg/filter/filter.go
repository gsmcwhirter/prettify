package filter

import (
	"fmt"
	"regexp"

	"github.com/tidwall/gjson"
)

var filterRegex = regexp.MustCompile(`^\.\.#\[.+\]$`)

// LineFilter is an interface for something that will deem lines allowable or not
type LineFilter interface {
	AllowBytes([]byte) bool
	Allow(string) bool
}

// lineFilter is an implementation of the LineFilter interface using gjson queries
type lineFilter struct {
	filters []string
}

// NewLineFilter constructs a struct implementing the LineFilter interface
//
// filters is a string of gjson queries. If an entry is not of the form ..#[%s], it will be turned into that form.
// A line needs to pass ALL of the filters to be allowed
func NewLineFilter(filters []string) LineFilter {
	lf := lineFilter{
		filters: filters,
	}

	// make sure all the filters have proper form
	for i := range lf.filters {
		if !filterRegex.MatchString(lf.filters[i]) {
			lf.filters[i] = fmt.Sprintf("..#[%s]", lf.filters[i])
		}
	}

	return lf
}

// AllowBytes determines if a byte-array line is allowed according to the filters
func (lf lineFilter) AllowBytes(line []byte) bool {
	for _, filter := range lf.filters {
		res := gjson.GetBytes(line, filter)
		if !res.Exists() {
			return false
		}
	}

	return true
}

// Allow determines if a string line is allowed according to the filters
func (lf lineFilter) Allow(line string) bool {
	for _, filter := range lf.filters {
		res := gjson.Get(line, filter)
		if !res.Exists() {
			return false
		}
	}

	return true
}
