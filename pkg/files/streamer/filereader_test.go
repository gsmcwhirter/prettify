package streamer

import (
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/pathutil"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
	"github.com/gsmcwhirter/prettify/pkg/testutil"
)

func Test_skipToEnd(t *testing.T) {
	type args struct {
		file io.Seeker
	}
	tests := []struct {
		name       string
		args       args
		wantEndPos int64
		wantErr    bool
	}{
		{
			name:       "basic test",
			args:       args{file: testutil.NewReadSeeker([]byte("testing"))},
			wantEndPos: 7,
			wantErr:    false,
		},
		{
			name:       "empty test",
			args:       args{file: testutil.NewReadSeeker([]byte{})},
			wantEndPos: 0,
			wantErr:    false,
		},
		{
			name:       "bad io",
			args:       args{file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{})},
			wantEndPos: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotEndPos, err := skipToEnd(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("skipToEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotEndPos != tt.wantEndPos {
				t.Errorf("skipToEnd() = %v, want %v", gotEndPos, tt.wantEndPos)
			}
		})
	}
}

func Test_peekEndPos(t *testing.T) {
	type args struct {
		file io.Seeker
	}
	tests := []struct {
		name        string
		args        args
		wantReadPos int64
		wantEndPos  int64
		wantErr     bool
	}{
		{
			name:        "basic test",
			args:        args{file: testutil.NewReadSeeker([]byte("testing"))},
			wantReadPos: 0,
			wantEndPos:  7,
			wantErr:     false,
		},
		{
			name:        "empty test",
			args:        args{file: testutil.NewReadSeeker([]byte{})},
			wantReadPos: 0,
			wantEndPos:  0,
			wantErr:     false,
		},
		{
			name:        "bad io",
			args:        args{file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 0})},
			wantReadPos: 0,
			wantEndPos:  0,
			wantErr:     true,
		},
		{
			name:        "bad io 2",
			args:        args{file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 1})},
			wantReadPos: 0,
			wantEndPos:  0,
			wantErr:     true,
		},
		{
			name:        "bad io 3",
			args:        args{file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 2})},
			wantReadPos: 1,
			wantEndPos:  1,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotEndPos, err := peekEndPos(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("peekEndPos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotEndPos != tt.wantEndPos {
				t.Errorf("peekEndPos() = %v, want %v", gotEndPos, tt.wantEndPos)
			}

			readPos, err := tt.args.file.Seek(0, io.SeekCurrent)
			if err != nil || readPos != tt.wantReadPos {
				t.Errorf("peekEndPos() read pointer = %v, want %v", readPos, tt.wantReadPos)
			}
		})
	}
}

func Test_forwardOneLine(t *testing.T) {
	type args struct {
		file io.ReadSeeker
	}
	tests := []struct {
		name     string
		args     args
		startPos int64
		wantPos  int64
		wantErr  bool
	}{
		{
			name: "start of file",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 0,
			wantPos:  10,
			wantErr:  false,
		},
		{
			name: "end of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 9,
			wantPos:  10,
			wantErr:  false,
		},
		{
			name: "start of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 10,
			wantPos:  20,
			wantErr:  false,
		},
		{
			name: "bad io",
			args: args{
				file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 2}),
			},
			startPos: 0,
			wantPos:  0,
			wantErr:  true,
		},
		{
			name: "bad io 2",
			args: args{
				file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 3}),
			},
			startPos: 0,
			wantPos:  0,
			wantErr:  true,
		},
		{
			name: "bad io 3",
			args: args{
				file: testutil.NewErrReadSeeker(0, 1, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 6}),
			},
			startPos: 0,
			wantPos:  0,
			wantErr:  true,
		},
		{
			name: "bad io 4",
			args: args{
				file: testutil.NewErrReadSeeker(1, 1, testutil.ErrReadSeekerConfig{ErrOnReadCall: 1, ErrOnSeekCall: 6, FillByte: byte('\n')}),
			},
			startPos: 0,
			wantPos:  1,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			gotPos, err := forwardOneLine(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("forwardOneLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPos != tt.wantPos {
				t.Errorf("forwardOneLine() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}

func Test_reverseOneLine(t *testing.T) {
	type args struct {
		file io.ReadSeeker
	}
	tests := []struct {
		name     string
		args     args
		startPos int64
		wantPos  int64
		wantErr  bool
	}{
		{
			name: "start of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 20,
			wantPos:  10,
			wantErr:  false,
		},
		{
			name: "end of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 19,
			wantPos:  10,
			wantErr:  false,
		},
		{
			name: "start of file",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos: 0,
			wantPos:  0,
			wantErr:  false,
		},
		{
			name: "bad io",
			args: args{
				file: testutil.NewErrReadSeeker(0, 30, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 2}),
			},
			startPos: 10,
			wantPos:  10,
			wantErr:  true,
		},
		{
			name: "bad io 2",
			args: args{
				file: testutil.NewErrReadSeeker(0, 10, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 3}),
			},
			startPos: 10,
			wantPos:  10,
			wantErr:  true,
		},
		{
			name: "bad io 3",
			args: args{
				file: testutil.NewErrReadSeeker(0, 10, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 4}),
			},
			startPos: 10,
			wantPos:  9,
			wantErr:  true,
		},
		{
			name: "bad io 4",
			args: args{
				file: testutil.NewErrReadSeeker(0, 10, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 5}),
			},
			startPos: 10,
			wantPos:  0,
			wantErr:  true,
		},
		{
			name: "bad io 5",
			args: args{
				file: testutil.NewErrReadSeeker(1, 10, testutil.ErrReadSeekerConfig{ErrOnReadCall: 1, ErrOnSeekCall: 5, FillByte: byte('\n')}),
			},
			startPos: 10,
			wantPos:  1,
			wantErr:  true,
		},
		{
			name: "bad io 6",
			args: args{
				file: testutil.NewErrReadSeeker(1, 10, testutil.ErrReadSeekerConfig{ErrOnReadCall: 1, ErrOnSeekCall: 5}),
			},
			startPos: 10,
			wantPos:  1,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			gotPos, err := reverseOneLine(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("reverseOneLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPos != tt.wantPos {
				t.Errorf("reverseOneLine() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}

func Test_readLineBefore(t *testing.T) {
	type args struct {
		file io.ReadSeeker
	}
	tests := []struct {
		name      string
		args      args
		startPos  int64
		wantSlice []byte
		wantPos   int64
		wantErr   bool
	}{
		{
			name: "start of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos:  20,
			wantSlice: []byte("testing 2\n"),
			wantPos:   10,
			wantErr:   false,
		},
		{
			name: "end of line",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos:  19,
			wantSlice: []byte("testing "),
			wantPos:   10,
			wantErr:   false,
		},
		{
			name: "continuation",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos:  10,
			wantSlice: []byte("testing 1\n"),
			wantPos:   0,
			wantErr:   false,
		},
		{
			name: "start of file",
			args: args{
				file: testutil.NewReadSeeker([]byte("testing 1\ntesting 2\n")),
			},
			startPos:  0,
			wantSlice: []byte{},
			wantPos:   0,
			wantErr:   false,
		},
		{
			name: "bad io",
			args: args{
				file: testutil.NewErrReadSeeker(0, 30, testutil.ErrReadSeekerConfig{ErrOnSeekCall: 2}),
			},
			startPos: 10,
			wantPos:  10,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			gotSlice, gotPos, err := readLineBefore(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("readLineBefore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSlice, tt.wantSlice) && (len(gotSlice) > 0 || len(tt.wantSlice) > 0) {
				t.Errorf("readLineBefore() gotSlice = %v, want %v", gotSlice, tt.wantSlice)
			}
			if gotPos != tt.wantPos {
				t.Errorf("readLineBefore() gotPos = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}

func Test_moveLines(t *testing.T) {
	type args struct {
		file      io.ReadSeeker
		skipLines int
	}
	tests := []struct {
		name           string
		args           args
		startPos       int64
		wantPos        int64
		wantLinesMoved int
		wantErr        bool
	}{
		{
			name: "forwards 1",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: 1,
			},
			startPos:       0,
			wantPos:        10,
			wantLinesMoved: 1,
			wantErr:        false,
		},
		{
			name: "forwards 2",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: 2,
			},
			startPos:       0,
			wantPos:        20,
			wantLinesMoved: 2,
			wantErr:        false,
		},
		{
			name: "forwards partial",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: 2,
			},
			startPos:       20,
			wantPos:        30,
			wantLinesMoved: 1,
			wantErr:        false,
		},
		{
			name: "backwards 1",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: -1,
			},
			startPos:       30,
			wantPos:        20,
			wantLinesMoved: -1,
			wantErr:        false,
		},
		{
			name: "backwards 2",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: -2,
			},
			startPos:       30,
			wantPos:        10,
			wantLinesMoved: -2,
			wantErr:        false,
		},
		{
			name: "backwards partial",
			args: args{
				file:      testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				skipLines: -2,
			},
			startPos:       10,
			wantPos:        0,
			wantLinesMoved: -1,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			linesMoved, err := moveLines(tt.args.file, tt.args.skipLines)
			if (err != nil) != tt.wantErr {
				t.Errorf("moveLines() error = %v, wantErr %v", err, tt.wantErr)
			}
			if linesMoved != tt.wantLinesMoved {
				t.Errorf("moveLines() linesMoved = %v, wantLinesMoved %v", linesMoved, tt.wantLinesMoved)
			}
			pos, err = tt.args.file.Seek(0, io.SeekCurrent)
			if err != nil || pos != tt.wantPos {
				t.Errorf("moveLines() post pos = %d, want %d (err = %v)", pos, tt.wantPos, err)
			}
		})
	}
}

func Test_catFile(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		file     io.ReadSeeker
		filename string
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		startPos  int64
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  0,
			wantPos:   30,
			wantErr:   false,
			wantBytes: []byte("testing 1\ntesting 2\ntesting 3\n"),
		},
		{
			name: "start at line 2",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  10,
			wantPos:   30,
			wantErr:   false,
			wantBytes: []byte("testing 2\ntesting 3\n"),
		},
		{
			name: "skip blank lines",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  0,
			wantPos:   33,
			wantErr:   false,
			wantBytes: []byte("testing 1\ntesting 2\ntesting 3\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  0,
			wantPos:   33,
			wantErr:   false,
			wantBytes: []byte("testing 1\n\n\ntesting 2\n\ntesting 3\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  0,
			wantPos:   33,
			wantErr:   false,
			wantBytes: []byte("test: testing 1\ntest: testing 2\ntest: testing 3\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  0,
			wantPos:   33,
			wantErr:   false,
			wantBytes: []byte("test: testing 1\ntest: \ntest: \ntest: testing 2\ntest: \ntest: testing 3\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			gotPos, err := catFile(context.Background(), tt.args.file, tt.args.filename, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("catFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPos != tt.wantPos {
				t.Errorf("catFile() = %v, want %v", gotPos, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func Test_tacFile(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		file     io.ReadSeeker
		filename string
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		startPos  int64
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  30,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("testing 3\ntesting 2\ntesting 1\n"),
		},
		{
			name: "start at line 2",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\ntesting 2\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  10,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("testing 1\n"),
		},
		{
			name: "skip blank lines",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  33,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("testing 3\ntesting 2\ntesting 1\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  33,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("testing 3\n\ntesting 2\n\n\ntesting 1\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  33,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("test: testing 3\ntest: testing 2\ntest: testing 1\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				file:     testutil.NewReadSeeker([]byte("testing 1\n\n\ntesting 2\n\ntesting 3\n")),
				filename: "test",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			startPos:  33,
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("test: testing 3\ntest: \ntest: testing 2\ntest: \ntest: \ntest: testing 1\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			pos, err := tt.args.file.Seek(tt.startPos, io.SeekStart)
			if err != nil || pos != tt.startPos {
				t.Errorf("Could not set file to desired test position (err = %v, got pos %d)", err, pos)
			} else {
				pos, err = tt.args.file.Seek(0, io.SeekCurrent)
				if err != nil || pos != tt.startPos {
					t.Errorf("Setting start position didn't stick (err = %v, got pos %d)", err, pos)
				}
			}

			gotPos, err := tacFile(context.Background(), tt.args.file, tt.args.filename, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("tacFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPos != tt.wantPos {
				t.Errorf("tacFile() = %v, want %v", gotPos, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("tacFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func TestCat(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		directory string
		filename  string
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("1\n2\n3\n"),
		},
		{
			name: "skip blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("1\n2\n3\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("1\n\n\n2\n\n3\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 1\ntest-out-3.log: 2\ntest-out-3.log: 3\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 1\ntest-out-3.log: \ntest-out-3.log: \ntest-out-3.log: 2\ntest-out-3.log: \ntest-out-3.log: 3\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			got, err := Cat(context.Background(), tt.args.directory, tt.args.filename, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPos {
				t.Errorf("Cat() pos = %v, want %v", got, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func TestCatFrom(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		directory string
		filename  string
		pos       int64
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       0,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("1\n2\n3\n"),
		},
		{
			name: "start from line 2",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       2,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("2\n3\n"),
		},
		{
			name: "skip blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       2,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("2\n3\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       2,
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("\n\n2\n\n3\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       2,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 2\ntest-out-3.log: 3\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				pos:       2,
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: \ntest-out-3.log: \ntest-out-3.log: 2\ntest-out-3.log: \ntest-out-3.log: 3\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			got, err := CatFrom(context.Background(), tt.args.directory, tt.args.filename, tt.args.pos, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("CatFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPos {
				t.Errorf("CatFrom() pos = %v, want %v", got, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func TestTac(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		directory string
		filename  string
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("3\n2\n1\n"),
		},
		{
			name: "skip blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("3\n2\n1\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("3\n\n2\n\n\n1\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 3\ntest-out-3.log: 2\ntest-out-3.log: 1\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   0,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 3\ntest-out-3.log: \ntest-out-3.log: 2\ntest-out-3.log: \ntest-out-3.log: \ntest-out-3.log: 1\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			got, err := Tac(context.Background(), tt.args.directory, tt.args.filename, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tac() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPos {
				t.Errorf("Tac() = %v, want %v", got, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func TestTail(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		directory string
		filename  string
		startLine int
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: 0,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("1\n2\n3\n"),
		},
		{
			name: "start from line 2",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: 1,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("2\n3\n"),
		},
		{
			name: "start from line -1",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: -1,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("3\n"),
		},
		{
			name: "start from line -2",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: -2,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("3\n"),
		},
		{
			name: "keep blank lines",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: -4,
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("\n2\n\n3\n"),
		},
		{
			name: "skip blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: -3,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: 2\ntest-out-3.log: 3\n"),
		},
		{
			name: "keep blank lines with filename",
			args: args{
				directory: pathutil.MustAbsPath("../testdata/taccat") + "/",
				filename:  "test-out-3.log",
				startLine: -5,
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-3.log: \ntest-out-3.log: \ntest-out-3.log: 2\ntest-out-3.log: \ntest-out-3.log: 3\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			got, err := Tail(context.Background(), tt.args.directory, tt.args.filename, tt.args.startLine, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPos {
				t.Errorf("Tail() pos = %v, want %v", got, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func TestTailFiles(t *testing.T) {
	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		filenames []string
		numLines  int
	}
	tests := []struct {
		name      string
		args      args
		lpArgs    lpArgs
		wantPos   int64
		wantErr   bool
		wantBytes []byte
	}{
		{
			name: "basic test",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: -3,
			},
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("2\n3\n"),
		},
		{
			name: "go back several files",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: -8,
			},
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			wantPos:   9,
			wantErr:   false,
			wantBytes: []byte("test-out-2.log: 8\ntest-out-2.log: 9\ntest-out-3.log: 1\ntest-out-3.log: \ntest-out-3.log: \ntest-out-3.log: 2\ntest-out-3.log: \ntest-out-3.log: 3\n"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			got, err := TailFiles(context.Background(), tt.args.filenames, tt.args.numLines, lp)
			if (err != nil) != tt.wantErr {
				t.Errorf("TailFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPos {
				t.Errorf("TailFiles() = %v, want %v", got, tt.wantPos)
			}
			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("catFile() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}

func Test_findStartingFile(t *testing.T) {
	type args struct {
		filenames []string
		numLines  int
	}
	tests := []struct {
		name             string
		args             args
		wantStartIndex   int
		wantStartLineNum int
		wantErr          bool
	}{
		{
			name: "last file negative lines",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: -1,
			},
			wantStartIndex:   3,
			wantStartLineNum: -1,
			wantErr:          false,
		},
		{
			name: "earlier file negative lines",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: -12,
			},
			wantStartIndex:   1,
			wantStartLineNum: -3,
			wantErr:          false,
		},
		{
			name: "first file positive lines",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: 2,
			},
			wantStartIndex:   3,
			wantStartLineNum: 2,
			wantErr:          false,
		},
		{
			name: "later file positive lines",
			args: args{
				filenames: []string{
					pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
					pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
				},
				numLines: 12,
			},
			wantStartIndex:   3,
			wantStartLineNum: 12,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotStartIndex, gotStartLineNum, err := findStartingFile(tt.args.filenames, tt.args.numLines)
			if (err != nil) != tt.wantErr {
				t.Errorf("findStartingFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStartIndex != tt.wantStartIndex {
				t.Errorf("findStartingFile() gotStartIndex = %v, want %v", gotStartIndex, tt.wantStartIndex)
			}
			if gotStartLineNum != tt.wantStartLineNum {
				t.Errorf("findStartingFile() gotStartLineNum = %v, want %v", gotStartLineNum, tt.wantStartLineNum)
			}
		})
	}
}
