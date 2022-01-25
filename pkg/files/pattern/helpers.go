package pattern

import (
	"fmt"
	"regexp"
	"strconv"
)

var dateCleanerRegex = regexp.MustCompile(`\D+`)

func cleanedDate(dateStr string) (intval uint64, strlen int) {
	cleanedDateStr := dateCleanerRegex.ReplaceAllString(dateStr, "")
	iVal, err := strconv.ParseInt(cleanedDateStr, 10, 64)
	if err != nil {
		return 0, 0
	}

	return uint64(iVal), len(cleanedDateStr)
}

// FilePatternsEqual is a helper function to determine if two Pattern objects
// are nearly equal (up to regex pointer)
func FilePatternsEqual(fp1, fp2 *Pattern) bool {
	if fp1 == nil && fp2 == nil {
		return true
	}

	if fp1 == nil || fp2 == nil {
		return false
	}

	if fp1.origPattern != fp2.origPattern {
		return false
	}

	if fp1.directory != fp2.directory {
		return false
	}

	if fp1.filePattern != fp2.filePattern {
		return false
	}

	if fp1.filenameGlob != fp2.filenameGlob {
		return false
	}

	return true
}

func formFileNameRegexp(fileName, requiredExtension string) *regexp.Regexp {
	if requiredExtension != "" {
		return regexp.MustCompile(fmt.Sprintf(`^%s-(out|error)(?:-(\d+))?(?:\.%s)$`, fileName, requiredExtension))
	}

	return regexp.MustCompile(fmt.Sprintf(`^%s-(out|error)(?:-(\d+))?(?:\..+)?$`, fileName))
}

func formFileNameGlob(fileName, requiredExtension string) string {
	if requiredExtension != "" {
		return fmt.Sprintf("%s-*.%s", fileName, requiredExtension)
	}

	return fmt.Sprintf("%s-*.*", fileName)
}
