package streamer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsmcwhirter/go-util/v9/deferutil"

	"github.com/gsmcwhirter/prettify/pkg/minmax"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
)

// NewlineByte is a contant containing the byte for a newline
const NewlineByte = byte('\n')

// MaxLineSizeBytes is a constant used for setting a file scanning buffer size
// We allow 1MB lines at maximum. Lines that are longer will be discarded / cause errors
const (
	MaxLineSizeBytes = 1 * 1024 * 1024 // 1 MB
	searchBufferSize = 64 * 1024       // 64 KB
)

var (
	searchBuffer = make([]byte, searchBufferSize)
	lineBuffer   = make([]byte, MaxLineSizeBytes+1)
)

func skipToEnd(file io.Seeker) (int64, error) {
	return file.Seek(0, io.SeekEnd) // start at the end
}

func peekEndPos(file io.Seeker) (int64, error) {
	pos, fErr := file.Seek(0, io.SeekCurrent)
	if fErr != nil {
		return 0, fErr
	}

	endPos, err := skipToEnd(file)
	if err == nil {
		_, err = file.Seek(pos, io.SeekStart)
	}
	return endPos, err
}

func reverseOneLine(file io.ReadSeeker) (pos int64, err error) {
	pos, err = file.Seek(0, io.SeekCurrent)
	if pos == 0 || err != nil {
		return pos, err
	}

	// assume we are at the start of a line. back up one character over the newline
	pos, err = file.Seek(-1, io.SeekCurrent)
	if err != nil {
		return pos, err
	}

	// debug print
	// fmt.Printf("step back pos %d\n", pos)

	var bytesRead int
	var seekBack int64
	for pos > 0 {
		seekBack = minmax.Int64Min(pos, searchBufferSize)

		// jump back to read a bit
		pos, err = file.Seek(-seekBack, io.SeekCurrent)
		if err != nil {
			return pos, err
		}

		// debug print
		// fmt.Printf("read at pos %d\n", pos)

		bytesRead, err = file.Read(searchBuffer)
		if err != nil && !errors.Is(err, io.EOF) {
			return pos, err
		}

		// debug print
		// fmt.Printf("read buffer %v\n", searchBuffer[:bytesRead])

		if bytesRead == 0 {
			return pos, nil
		}

		for i := minmax.IntMin(int(seekBack), bytesRead) - 1; i >= 0; i-- {
			if searchBuffer[i] == NewlineByte {
				return file.Seek(pos+int64(i)+1, io.SeekStart)
			}
		}

		// debug block
		// pos, err = file.Seek(0, io.SeekCurrent)
		// if err != nil {
		//	return
		// }
		// fmt.Printf("curr pos %d\n", pos)

		pos, err = file.Seek(-int64(bytesRead), io.SeekCurrent)
		if err != nil {
			return pos, err
		}
	}

	return pos, nil
}

func forwardOneLine(file io.ReadSeeker) (int64, error) {
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return pos, err
	}

	endPos, fErr := peekEndPos(file)
	if fErr != nil {
		return pos, fErr
	}

	// debug print
	// fmt.Printf("start pos %d\n", pos)

	if pos == endPos {
		return pos, nil
	}

	for pos < endPos {
		// read the character we are at
		// debug print
		// fmt.Printf("read at pos %d\n", pos)

		bytesRead, readErr := file.Read(searchBuffer)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return pos, readErr
		}

		if bytesRead == 0 {
			return pos, nil
		}

		// debug print
		// fmt.Printf("read buffer %v\n", searchBuffer[:bytesRead])

		for i := 0; i < bytesRead; i++ {
			if searchBuffer[i] == NewlineByte {
				pos = pos + int64(i) + 1

				// debug print
				// fmt.Printf("Found newline at i=%d, pos=%d\n", i, pos)

				pos, err = file.Seek(pos, io.SeekStart)
				return pos, err
			}
		}

		// debug block
		// pos, err = file.Seek(0, io.SeekCurrent)
		// if err != nil {
		//	return
		// }
		// fmt.Printf("curr pos %d\n", pos)

		pos += int64(bytesRead)
	}

	return pos, nil
}

func readLineBefore(file io.ReadSeeker) ([]byte, int64, error) {
	// note: assumes that the file read pointer is just after a newline character.
	// and attempts to read the line on the other side of that newline

	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return []byte{}, pos, err
	}

	if pos == 0 {
		return []byte{}, pos, nil
	}

	// debug print
	// fmt.Printf("reading from %d\n", pos)

	readTarget := minmax.Int64Max(pos-MaxLineSizeBytes, 0)

	_, err = file.Seek(readTarget, io.SeekStart)
	if err != nil {
		return []byte{}, pos, err
	}

	numBytes, readErr := file.Read(lineBuffer)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return []byte{}, pos, readErr
	}

	// debug print
	// fmt.Printf("numBytes %d, buffer slice %v, buffer str '%s'\n", numBytes, lineBuffer[:numBytes], string(lineBuffer[:numBytes]))

	if numBytes < 2 {
		pos -= int64(numBytes)
		return []byte{}, pos, nil
	}

	endByte := readTarget + int64(numBytes)
	skipBytes := endByte - pos

	// debug print
	// fmt.Printf("skip bytes %d\n", skipBytes)

	lastIndex := numBytes - int(skipBytes) - 1

	// debug print
	// fmt.Printf("lastIndex %d\n", lastIndex)

	pos-- // the newline we were just at
	// -1 because we should have just been at a newline
	for i := lastIndex - 1; i >= 0; i-- {
		if i == 0 {
			var slice []byte
			pos = 0
			if lineBuffer[lastIndex] == NewlineByte {
				slice = lineBuffer[i : lastIndex+1]
			} else {
				slice = lineBuffer[i:lastIndex]
			}

			_, err = file.Seek(pos, io.SeekStart)

			// debug print
			// fmt.Printf("returning i=0 pos %d slice %v\n", pos, slice)
			return slice, pos, err
		} else if lineBuffer[i] == NewlineByte {
			var slice []byte
			if lineBuffer[lastIndex] == NewlineByte {
				slice = lineBuffer[i+1 : lastIndex+1]
			} else {
				slice = lineBuffer[i+1 : lastIndex]
			}
			_, err = file.Seek(pos, io.SeekStart)

			// debug print
			// fmt.Printf("returning i!=0 pos %d slice %v\n", pos, slice)
			return slice, pos, err
		}
		pos--
	}

	return []byte{}, pos, nil
}

func moveLines(file io.ReadSeeker, skipLines int) (linesMoved int, err error) {
	// Skip to the end of the file and set up to read only abs(startLine) many lines
	var pos int64 = 1
	var endPos int64

	switch {
	case skipLines == 0:
		return 0, nil
	case skipLines < 0:
		for pos > 0 && skipLines < 0 {
			pos, err = reverseOneLine(file)
			if err != nil {
				return linesMoved, err
			}

			skipLines++
			linesMoved--
		}
	default:
		endPos, err = peekEndPos(file)
		if err != nil {
			return linesMoved, err
		}

		for pos < endPos && skipLines > 0 {
			pos, err = forwardOneLine(file)
			if err != nil {
				return linesMoved, err
			}

			skipLines--
			linesMoved++
		}
	}

	return linesMoved, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// ScanLines is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func scanLinesWithNewline(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, append(dropCR(data[0:i]), 10), nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}

	// Request more data.
	return 0, nil, nil
}

func catFile(ctx context.Context, file io.ReadSeeker, filename string, lp linehandler.LineHandler) (int64, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(scanLinesWithNewline)
	buffer := make([]byte, MaxLineSizeBytes+1)
	scanner.Buffer(buffer, MaxLineSizeBytes)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			pos, err := file.Seek(0, io.SeekCurrent)
			if err != nil {
				return 0, err
			}
			err = ctx.Err()
			return pos, err
		default:
		}

		line := scanner.Text() // with newline
		lp.HandleLine(filename, line)
	}

	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	return pos, nil
}

// func catGzip(file io.Reader, filename string, lp linehandler.LineHandler) error {
//	scanner := bufio.NewScanner(file)
//	for scanner.Scan() {
//		line := scanner.Text()
//		lp.HandleLine(filename, line)
//	}

//	return nil
// }

// Cat a tsar log file, possibly ignoring some lines
func Cat(ctx context.Context, directory, filename string, lp linehandler.LineHandler) (int64, error) {
	// directory and filename should come from filepath.Split()
	return CatFrom(ctx, directory, filename, 0, lp)
}

// CatFrom cats a file starting from the specified byte offset
func CatFrom(ctx context.Context, directory, filename string, pos int64, lp linehandler.LineHandler) (int64, error) {
	// debug print
	// fmt.Printf("Catting %s%s from %d\n", directory, filename, pos)

	file, err := os.Open(directory + filename)
	if err != nil {
		return 0, err
	}
	defer deferutil.CheckDefer(file.Close)

	if strings.HasSuffix(filename, ".gz") {
		// debug print
		// fmt.Println("Opening a gzip wrapper around the file")

		// if pos == 0 {
		//	gzf, gzerr := gzip.NewReader(file)
		//	if gzerr != nil {
		//		return 0, gzerr
		//	}
		//	defer deferutil.MustClose(gzf)

		//	return 0, catGzip(gzf, filename, lp)
		// }

		fmt.Fprintln(os.Stderr, "Cannot specify a pos to CatFrom with a gzip file. Treating as a normal file (probably with bad results)")
	}

	// debug print
	// fmt.Println("No gzip wrapper required")

	// not gzip
	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return catFile(ctx, file, filename, lp)
}

// Tail will tail a file, possibly ignoring some lines
func Tail(ctx context.Context, directory, filename string, skipLines int, lp linehandler.LineHandler) (int64, error) {
	// directory and filename should come from filepath.Split()
	file, err := os.Open(directory + filename)
	if err != nil {
		return 0, err
	}
	defer deferutil.CheckDefer(file.Close)

	if skipLines < 0 {
		pos, serr := skipToEnd(file)
		if serr != nil {
			return pos, serr
		}
	}

	_, err = moveLines(file, skipLines)
	if err != nil {
		return 0, err
	}

	return catFile(ctx, file, filename, lp)
}

func checkFileForStartingFile(fname string, remainingLines int) (int, error) {
	file, err := os.Open(fname)
	if err != nil {
		return 0, err
	}
	defer deferutil.CheckDefer(file.Close)

	if remainingLines < 0 {
		_, err = skipToEnd(file)
		if err != nil {
			return 0, err
		}
	}

	linesMoved, err := moveLines(file, remainingLines)
	if err != nil {
		return remainingLines, err
	}

	return remainingLines - linesMoved, nil
}

func findStartingFile(filenames []string, remainingLines int) (startIndex, startLineNum int, err error) {
	for i := len(filenames) - 1; i >= 0; i-- {
		startIndex = i
		startLineNum = remainingLines

		remainingLines, err = checkFileForStartingFile(filenames[i], remainingLines)
		if err != nil {
			return startIndex, startLineNum, err
		}

		if remainingLines >= 0 {
			return startIndex, startLineNum, err
		}
	}

	return startIndex, startLineNum, err
}

func tailFileCat(ctx context.Context, fname string, startLineNum int, lp linehandler.LineHandler) (int64, error) {
	file, err := os.Open(fname)
	if err != nil {
		return 0, err
	}
	defer deferutil.CheckDefer(file.Close)

	_, fileName := filepath.Split(fname)

	if startLineNum < 0 {
		_, err = skipToEnd(file)
		if err != nil {
			return 0, err
		}
	}

	_, err = moveLines(file, startLineNum)
	if err != nil {
		return 0, err
	}

	endFilePos, err := catFile(ctx, file, fileName, lp)
	if err != nil {
		return 0, err
	}

	return endFilePos, nil
}

// TailFiles gets the last numLines lines from the files
func TailFiles(ctx context.Context, filenames []string, numLines int, lp linehandler.LineHandler) (int64, error) {
	if len(filenames) == 0 {
		return 0, errors.New("cannot TailFiles on an empty list of filenames")
	}

	startIndex, startLineNum, err := findStartingFile(filenames, numLines)
	if err != nil {
		return 0, err
	}

	var endFilePos int64
	for i := startIndex; i < len(filenames); i++ {
		if i == startIndex {
			endFilePos, err = tailFileCat(ctx, filenames[i], startLineNum, lp)
		} else {
			endFilePos, err = tailFileCat(ctx, filenames[i], 0, lp)
		}

		if err != nil {
			return 0, err
		}
	}

	return endFilePos, nil
}

func tacFile(ctx context.Context, file io.ReadSeeker, filename string, lp linehandler.LineHandler) (int64, error) {
	// debug print
	// fmt.Printf("taccing %s\n", filename)

	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return pos, err
	}

	// debug print
	// fmt.Printf("pos post skip %d\n", pos)

	for pos > 0 {
		select {
		case <-ctx.Done():
			return pos, ctx.Err()
		default:
		}

		lineBytes, newPos, readErr := readLineBefore(file)
		if readErr != nil {
			return newPos, readErr
		}

		pos = newPos

		// debug print
		// fmt.Printf("lineBytes %v\n", lineBytes)

		line := string(lineBytes)
		lp.HandleLine(filename, line)

		// // debug print
		// // fmt.Printf("pos after print %d\n", pos)
	}

	return pos, nil
}

// Tac prints out contents of a file backwards
func Tac(ctx context.Context, directory, filename string, lp linehandler.LineHandler) (int64, error) {
	// directory and filename should come from filepath.Split()
	// debug print
	// fmt.Printf("Catting %s%s from %d\n", directory, filename, pos)

	file, err := os.Open(directory + filename)
	if err != nil {
		return 0, err
	}
	defer deferutil.CheckDefer(file.Close)

	if strings.HasSuffix(filename, ".gz") {
		fmt.Fprintln(os.Stderr, "Cannot Tac a gzip file. Treating as a normal file (probably with bad results)")
	}

	pos, err := skipToEnd(file)
	if err != nil {
		return pos, err
	}

	return tacFile(ctx, file, filename, lp)
}
