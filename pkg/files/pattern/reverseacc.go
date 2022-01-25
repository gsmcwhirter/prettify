package pattern

import "os"

type reverseAccumulatorRecord struct {
	path string
	info os.FileInfo
}

type reverseAccumulator struct {
	Paths []reverseAccumulatorRecord
}

func (ra *reverseAccumulator) record(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	ra.Paths = append(ra.Paths, reverseAccumulatorRecord{
		path: path,
		info: info,
	})

	return nil
}

// Sorting interface
func (ra *reverseAccumulator) Len() int {
	return len(ra.Paths)
}

// Sorting interface
func (ra *reverseAccumulator) Swap(i, j int) {
	ra.Paths[i], ra.Paths[j] = ra.Paths[j], ra.Paths[i]
}

// Sorting interface
func (ra *reverseAccumulator) Less(i, j int) bool {
	if len(ra.Paths) == 0 {
		return false
	}

	return ra.Paths[j].path < ra.Paths[i].path
}
