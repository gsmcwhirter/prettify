package pattern

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_cleanedDate(t *testing.T) {
	t.Parallel()
	type args struct {
		dateStr string
	}
	tests := []struct {
		name       string
		args       args
		wantIntval uint64
		wantStrlen int
	}{
		{
			name:       "just ints",
			args:       args{dateStr: "1234567890"},
			wantIntval: 1234567890,
			wantStrlen: 10,
		},
		{
			name:       "other chars",
			args:       args{dateStr: "2018-10-01 at 17:30"},
			wantIntval: 201810011730,
			wantStrlen: 12,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotIntval, gotStrlen := cleanedDate(tt.args.dateStr)
			if gotIntval != tt.wantIntval {
				t.Errorf("cleanedDate() gotIntval = %v, want %v", gotIntval, tt.wantIntval)
			}
			if gotStrlen != tt.wantStrlen {
				t.Errorf("cleanedDate() gotStrlen = %v, want %v", gotStrlen, tt.wantStrlen)
			}
		})
	}
}

func Test_FilePatternsEqual(t *testing.T) {
	t.Parallel()
	type args struct {
		fp1 *Pattern
		fp2 *Pattern
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal nil regex",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
			},
			want: true,
		},
		{
			name: "equal despite regex",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: regexp.MustCompile(`f[o0]o`),
				},
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: regexp.MustCompile(`b[a4]r`),
				},
			},
			want: true,
		},
		{
			name: "diff origPattern",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: &Pattern{
					origPattern:   "foo2",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
			},
			want: false,
		},
		{
			name: "diff directory",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar2",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
			},
			want: false,
		},
		{
			name: "diff filePattern",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz2",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
			},
			want: false,
		},
		{
			name: "diff filenameGlob",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux2",
					filenameRegex: nil,
				},
			},
			want: false,
		},
		{
			name: "both nil",
			args: args{
				fp1: nil,
				fp2: nil,
			},
			want: true,
		},
		{
			name: "first nil",
			args: args{
				fp1: nil,
				fp2: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
			},
			want: false,
		},
		{
			name: "second nil",
			args: args{
				fp1: &Pattern{
					origPattern:   "foo",
					directory:     "bar",
					filePattern:   "baz",
					filenameGlob:  "quux",
					filenameRegex: nil,
				},
				fp2: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := FilePatternsEqual(tt.args.fp1, tt.args.fp2); got != tt.want {
				t.Errorf("FilePatternsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formFileNameRegexp(t *testing.T) {
	t.Parallel()
	type args struct {
		fileName          string
		requiredExtension string
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
		{
			name: "without extension",
			args: args{
				fileName:          "foo",
				requiredExtension: "",
			},
			want: regexp.MustCompile(`^foo-(out|error)(?:-(\d+))?(?:\..+)?$`),
		},
		{
			name: "with extension",
			args: args{
				fileName:          "foo",
				requiredExtension: "test",
			},
			want: regexp.MustCompile(`^foo-(out|error)(?:-(\d+))?(?:\.test)$`),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formFileNameRegexp(tt.args.fileName, tt.args.requiredExtension); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formFileNameRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formFileNameGlob(t *testing.T) {
	t.Parallel()
	type args struct {
		fileName          string
		requiredExtension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "without extension",
			args: args{
				fileName:          "foo",
				requiredExtension: "",
			},
			want: "foo-*.*",
		},
		{
			name: "with extension",
			args: args{
				fileName:          "foo",
				requiredExtension: "test",
			},
			want: "foo-*.test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formFileNameGlob(tt.args.fileName, tt.args.requiredExtension); got != tt.want {
				t.Errorf("formFileNameGlob() = %v, want %v", got, tt.want)
			}
		})
	}
}
