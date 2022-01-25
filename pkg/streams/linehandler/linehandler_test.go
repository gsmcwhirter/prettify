package linehandler

import (
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/testutil"
)

func TestNewLinePrinter(t *testing.T) {
	t.Parallel()
	expected := linePrinter{
		withPath:     "a",
		withPretty:   true,
		withColor:    true,
		withSort:     false,
		withBlanks:   true,
		withFilename: true,
		printf:       nil,
	}

	type args struct {
		withBlanks   bool
		withFilename bool
		withPath     string
		withPretty   bool
		withColor    bool
		withSort     bool
	}
	tests := []struct {
		name string
		args args
		want *linePrinter
	}{
		{
			name: "basic test",
			args: args{
				withBlanks:   true,
				withFilename: true,
				withPath:     "a",
				withPretty:   true,
				withColor:    true,
				withSort:     false,
			},
			want: &expected,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewLinePrinter(Options{
				WithBlanks:   tt.args.withBlanks,
				WithFilename: tt.args.withFilename,
				JSONPath:     tt.args.withPath,
				Pretty:       tt.args.withPretty,
				Color:        tt.args.withColor,
				Sort:         tt.args.withSort,
				Printf:       nil,
			})

			if got.(*linePrinter).printf == nil {
				t.Errorf("NewLinePrinter() didn't set printf default")
			}

			got.(*linePrinter).printf = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLinePrinter() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLinePrinter_HandleLine(t *testing.T) {
	t.Parallel()

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		filename string
		line     string
	}
	tests := []struct {
		name      string
		lpArgs    lpArgs
		args      args
		wantBytes []byte
	}{
		{
			name: "default out",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "foo",
			},
			wantBytes: []byte("foo"),
		},
		{
			name: "without blanks",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "",
			},
			wantBytes: []byte{},
		},
		{
			name: "with blanks",
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "",
			},
			wantBytes: []byte(""),
		},
		{
			name: "with filename noblank",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "",
			},
			wantBytes: []byte{},
		},
		{
			name: "with filename blank",
			lpArgs: lpArgs{
				withBlanks:   true,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "",
			},
			wantBytes: []byte("test: "),
		},
		{
			name: "with filename noblank normal",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     "foo",
			},
			wantBytes: []byte("test: foo"),
		},
		{
			name: "pretty no color",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   true,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     `{"b": 1, "a": "foo"}`,
			},
			wantBytes: []byte("{\n  \"b\": 1,\n  \"a\": \"foo\"\n}"),
		},
		{
			name: "pretty color",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   true,
				withColor:    true,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     `{"b": 1, "a": "foo"}`,
			},
			wantBytes: []byte("{\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d\n}"),
		},
		{
			name: "pretty color filename",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   true,
				withColor:    true,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				filename: "test",
				line:     `{"b": 1, "a": "foo"}`,
			},
			wantBytes: []byte("test: {\n  \x1b\x5b\x39\x34\x6d\"b\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x33\x6d1\x1b\x5b\x30\x6d,\n  \x1b\x5b\x39\x34\x6d\"a\"\x1b\x5b\x30\x6d: \x1b\x5b\x39\x32\x6d\"foo\"\x1b\x5b\x30\x6d\n}"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start
			lp := NewLinePrinter(Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			lp.HandleLine(tt.args.filename, tt.args.line)

			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("HandleLine() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}
