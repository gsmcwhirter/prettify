package finder

import (
	"os"
	"strings"
)

func cleanDirName(dir string) string {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	return dir
}

func dirExists(dir string) bool {
	finfo, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return finfo.IsDir()
}
