package formatter

import (
	"bytes"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

func init() {
	gjson.DisableModifiers = true
}

type gJSONOutputType int

const (
	space gJSONOutputType = iota
	csv
	tsv
	nlsv
)

var (
	commaBytes   = []byte(",")
	tabBytes     = []byte("\t")
	spaceBytes   = []byte(" ")
	newlineBytes = []byte("\n")
)

func formatResultStrings(resStrings []string, separatorType gJSONOutputType) string {
	switch separatorType {
	case csv:
		return strings.Join(resStrings, ",")
	case tsv:
		return strings.Join(resStrings, "\t")
	case space:
		return strings.Join(resStrings, " ")
	default:
		return strings.Join(resStrings, "\n")
	}
}

func formatResultBytes(resBytes [][]byte, separatorType gJSONOutputType) []byte {
	switch separatorType {
	case csv:
		return bytes.Join(resBytes, commaBytes)
	case tsv:
		return bytes.Join(resBytes, tabBytes)
	case space:
		return bytes.Join(resBytes, spaceBytes)
	default:
		return bytes.Join(resBytes, newlineBytes)
	}
}

func getGJSONPaths(formatSelectors string) ([]string, gJSONOutputType) {
	rawGJSONPaths := strings.Split(formatSelectors, ",")
	gJSONPaths := make([]string, 0)
	separatorType := nlsv

	for _, component := range rawGJSONPaths {
		component = strings.TrimSpace(component)

		switch strings.ToLower(component) {
		case "|@ssv":
			separatorType = space
		case "|@csv":
			separatorType = csv
		case "|@tsv":
			separatorType = tsv
		case "|@nlsv":
			separatorType = nlsv
		default:
			gJSONPaths = append(gJSONPaths, component)
		}
	}

	return gJSONPaths, separatorType
}

func gjsonResultBytesToByteArray(line []byte, res gjson.Result) []byte {
	if res.Type == gjson.JSON {
		return []byte(res.Raw)
	}

	if res.Index <= 0 {
		return []byte(res.Raw)
	}

	if res.Type == gjson.String {
		return []byte(res.String())
	}

	return line[res.Index : res.Index+len(res.Raw)]

	// if res.Type == gjson.String { // strip off quotes
	//	return line[res.Index+1 : res.Index+len(res.Raw)-1]
	// }

	// return line[res.Index : res.Index+len(res.Raw)]
}

// FormatLine runs the formatSelectors (csv gjson selectors and an optional custom separator indicator)
// against the line and reformats the data into the requested format.
func FormatLine(line, formatSelectors string, prettyFmt, withColor, sortKeys bool) string {
	gJSONPaths, separatorType := getGJSONPaths(formatSelectors)

	resList := gjson.GetMany(line, gJSONPaths...)
	resStrings := make([]string, len(resList))
	for i := 0; i < len(resList); i++ {
		res := resList[i]
		switch {
		case !res.Exists():
			resStrings[i] = ""
		case res.Type == gjson.String:
			resStrings[i] = res.String()
		default:
			var toPrint string
			if prettyFmt {
				toPrint = PrettyLine(res.String(), withColor, sortKeys)
			} else {
				toPrintBytes := []byte(res.String())
				toPrintBytes = UglyLineBytes(toPrintBytes, withColor, sortKeys)
				toPrint = string(toPrintBytes)
			}

			toPrint = strings.TrimRight(toPrint, "\n")
			resStrings[i] = toPrint
		}
	}

	return formatResultStrings(resStrings, separatorType)
}

// FormatLineBytes runs the formatSelectors (csv gjson selectors and an optional custom separator indicator)
// against the line and reformats the data into the requested format.
func FormatLineBytes(line []byte, formatSelectors string, prettyFmt, withColor, sortKeys bool) []byte {
	gJSONPaths, separatorType := getGJSONPaths(formatSelectors)

	resList := gjson.GetManyBytes(line, gJSONPaths...)
	resSlices := make([][]byte, len(resList))
	for i := 0; i < len(resList); i++ {
		res := resList[i]
		if !res.Exists() {
			resSlices[i] = line[0:0]
		} else {
			resBytes := gjsonResultBytesToByteArray(line, res)

			if res.Type == gjson.JSON {
				if prettyFmt {
					resBytes = PrettyLineBytes(resBytes, withColor, sortKeys)
				} else {
					resBytes = UglyLineBytes(resBytes, withColor, sortKeys)
				}
			}

			resSlices[i] = resBytes
		}
	}

	return formatResultBytes(resSlices, separatorType)
}

// PrettyLine turns a json line into pretty form (one field per line, etc), possibly with sorted keys and colorized
func PrettyLine(line string, withColor, sortKeys bool) string {
	toPrintBytes := []byte(line)
	opts := pretty.Options{
		Width:    pretty.DefaultOptions.Width,
		Prefix:   pretty.DefaultOptions.Prefix,
		Indent:   pretty.DefaultOptions.Indent,
		SortKeys: sortKeys,
	}
	toPrintBytes = pretty.PrettyOptions(toPrintBytes, &opts)

	if withColor {
		toPrintBytes = pretty.Color(toPrintBytes, pretty.TerminalStyle)
	}

	return string(toPrintBytes)
}

// PrettyLineBytes turns a json line into pretty form (one field per line, etc), possibly with sorted keys and colorized
func PrettyLineBytes(line []byte, withColor, sortKeys bool) []byte {
	opts := pretty.Options{
		Width:    pretty.DefaultOptions.Width,
		Prefix:   pretty.DefaultOptions.Prefix,
		Indent:   pretty.DefaultOptions.Indent,
		SortKeys: sortKeys,
	}
	output := pretty.PrettyOptions(line, &opts)

	if withColor {
		output = pretty.Color(output, pretty.TerminalStyle)
	}

	return output
}

// UglyLineBytes turns a json line(s) into compact form, possibly with sorted keys and colorized
func UglyLineBytes(line []byte, withColor, sortKeys bool) []byte {
	if sortKeys {
		line = PrettyLineBytes(line, false, sortKeys)
	}

	line = pretty.Ugly(line)
	if withColor {
		line = pretty.Color(line, pretty.TerminalStyle)
	}

	return line
}
