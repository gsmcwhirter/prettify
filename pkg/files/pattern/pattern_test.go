package pattern

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/pathutil"
)

func TestNewFilePattern(t *testing.T) {
	t.Parallel()
	type args struct {
		pat string
	}
	tests := []struct {
		name   string
		args   args
		wantFp Pattern
	}{
		{
			name: "clean name",
			args: args{
				pat: "./foo/bar",
			},
			wantFp: Pattern{
				origPattern:  "./foo/bar",
				directory:    pathutil.MustAbsPath("./foo/") + "/",
				filePattern:  "bar",
				filenameGlob: "bar-*.*",
			},
		},
		{
			name: "with .log",
			args: args{
				pat: "./foo/bar.log",
			},
			wantFp: Pattern{
				origPattern:  "./foo/bar.log",
				directory:    pathutil.MustAbsPath("./foo/") + "/",
				filePattern:  "bar",
				filenameGlob: "bar-*.log",
			},
		},
		{
			name: "with trailing -",
			args: args{
				pat: "./foo/bar---",
			},
			wantFp: Pattern{
				origPattern:  "./foo/bar---",
				directory:    pathutil.MustAbsPath("./foo/") + "/",
				filePattern:  "bar",
				filenameGlob: "bar-*.*",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotFp := NewPattern(tt.args.pat); !FilePatternsEqual(&gotFp, &tt.wantFp) {
				t.Errorf("NewPattern() = %v, want %v", gotFp, tt.wantFp)
			}

			if gotFp := NewPatternPtr(tt.args.pat); !FilePatternsEqual(gotFp, &tt.wantFp) {
				t.Errorf("NewPatternPtr() = %v, want %v", *gotFp, tt.wantFp)
			}

			if gotFp := NewPatternPtrWithOriginal(tt.args.pat, tt.args.pat); !FilePatternsEqual(gotFp, &tt.wantFp) {
				t.Errorf("NewPatternPtrWithOriginal() = %v, want %v", gotFp, tt.wantFp)
			}
		})
	}
}

func TestFilePattern_matchesFile(t *testing.T) {
	t.Parallel()
	type args struct {
		dirName  string
		fileName string
	}
	tests := []struct {
		name string
		fp   *Pattern
		args args
		want bool
	}{
		{
			name: "matches",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"bar-error-1.log",
			},
			want: true,
		},
		{
			name: "matches in subdirectory",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/baz/") + "/",
				"bar-out-1.log",
			},
			want: true,
		},
		{
			name: "fails for wrong filename",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"baz-out-1.log",
			},
			want: false,
		},
		{
			name: "fails for bad filename",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"bar-baz-error-1.log",
			},
			want: false,
		},
		{
			name: "okay with longer filename",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"bar-out-1.log.log",
			},
			want: true,
		},
		{
			name: "okay with non-log extension",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"bar-out-1.jpl",
			},
			want: true,
		},
		{
			name: "fails for no out/error",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"bar-abcd.log",
			},
			want: false,
		},
		{
			name: "fails for wrong directory",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/../") + "/",
				"bar-out-1.log",
			},
			want: false,
		},
		{
			name: "fails for prefix",
			fp: &Pattern{
				origPattern:   "./testdata/foo/bar",
				directory:     pathutil.MustAbsPath("./testdata/foo/") + "/",
				filePattern:   "bar",
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			args: args{
				pathutil.MustAbsPath("./testdata/foo/") + "/",
				"baz-bar-out-1.log",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fp.matchesFile(tt.args.dirName, tt.args.fileName); got != tt.want {
				t.Errorf("FilePattern.matchesFile(%s) = %v, want %v", tt.args.fileName, got, tt.want)
			}
		})
	}
}

func TestFilePattern_Pattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fp   *Pattern
		want string
	}{
		{
			name: "basic test",
			fp: &Pattern{
				origPattern: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fp.Pattern(); got != tt.want {
				t.Errorf("FilePattern.Pattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePattern_Which(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fp   *Pattern
		want string
	}{
		{
			name: "basic test",
			fp: &Pattern{
				filenameGlob: "glob",
				directory:    "test/",
			},
			want: "test/glob",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fp.Which(); got != tt.want {
				t.Errorf("FilePattern.Which() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePattern_WalkFiles(t *testing.T) {
	t.Parallel()
	type args struct {
		walkFunc filepath.WalkFunc
	}
	tests := []struct {
		name        string
		fp          *Pattern
		args        args
		wantRecords []walkRecorderRecord
		wantErr     bool
	}{
		{
			name: "match one no filters",
			fp: &Pattern{
				origPattern:   "foo",
				directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
				filePattern:   "foo",
				filenameGlob:  formFileNameGlob("foo", ""),
				filenameRegex: formFileNameRegexp("foo", ""),
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl", nil},
			},
			wantErr: false,
		},
		{
			name: "match several no filters",
			fp: &Pattern{
				origPattern:   "bar",
				directory:     pathutil.MustAbsPath("../testdata/foo") + "/",
				filePattern:   "bar",
				filenameGlob:  formFileNameGlob("bar", ""),
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-error.log", nil},
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-out.log", nil},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := walkRecorder{[]walkRecorderRecord{}}
			tt.args = args{recorder.Record}
			if err := tt.fp.WalkFiles(context.Background(), tt.args.walkFunc); (err != nil) != tt.wantErr {
				t.Errorf("FilePattern.WalkFiles() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(recorder.Records, tt.wantRecords) && (len(recorder.Records) > 0 || len(tt.wantRecords) > 0) {
				t.Errorf("FilePattern.WalkFiles() records = %v, want %v", recorder.Records, tt.wantRecords)
			}
		})
	}
}

func TestFilePattern_WalkFilesReverse(t *testing.T) {
	t.Parallel()
	type args struct {
		walkFunc filepath.WalkFunc
	}
	tests := []struct {
		name        string
		fp          *Pattern
		args        args
		wantRecords []walkRecorderRecord
		wantErr     bool
	}{
		{
			name: "match one no filters",
			fp: &Pattern{
				origPattern:   "foo",
				directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
				filePattern:   "foo",
				filenameGlob:  formFileNameGlob("foo", ""),
				filenameRegex: formFileNameRegexp("foo", ""),
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl", nil},
			},
			wantErr: false,
		},
		{
			name: "match several no filters",
			fp: &Pattern{
				origPattern:   "bar",
				directory:     pathutil.MustAbsPath("../testdata/foo") + "/",
				filePattern:   "bar",
				filenameGlob:  formFileNameGlob("bar", ""),
				filenameRegex: formFileNameRegexp("bar", ""),
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-out.log", nil},
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-error.log", nil},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := walkRecorder{[]walkRecorderRecord{}}
			tt.args = args{recorder.Record}
			if err := tt.fp.WalkFilesReverse(context.Background(), tt.args.walkFunc); (err != nil) != tt.wantErr {
				t.Errorf("FilePattern.WalkFilesReverse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(recorder.Records, tt.wantRecords) && (len(recorder.Records) > 0 || len(tt.wantRecords) > 0) {
				t.Errorf("FilePattern.WalkFilesReverse() records = %v, want %v", recorder.Records, tt.wantRecords)
			}
		})
	}
}
