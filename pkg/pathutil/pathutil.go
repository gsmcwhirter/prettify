package pathutil

import "path/filepath"

// MustAbsPath returns the absolute path for a path or panics
func MustAbsPath(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return path
}
