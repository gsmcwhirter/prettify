package finder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
)

// Test to make sure that the searchDirs are being normalized properly
func TestNewFinder(t *testing.T) {
	t.Parallel()
	type args struct {
		searchDirs []string
	}
	tests := []struct {
		name string
		args args
		want *Finder
	}{
		{
			name: "searchDirs with trailing /",
			args: args{
				searchDirs: []string{
					"../testdata/foo/",
					"../testdata/bar/",
				},
			},
			want: &Finder{
				SearchDirectories: []string{
					"../testdata/foo/",
					"../testdata/bar/",
				},
			},
		},
		{
			name: "searchDirs without trailing /",
			args: args{
				searchDirs: []string{
					"../testdata/foo",
					"../testdata/bar",
				},
			},
			want: &Finder{
				SearchDirectories: []string{
					"../testdata/foo/",
					"../testdata/bar/",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewFinder(tt.args.searchDirs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFinder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinder_markSeen(t *testing.T) {
	t.Parallel()
	type fields struct {
		SearchDirectories []string
		foundFiles        bool
	}
	type args struct {
		in0 string
		in1 os.FileInfo
		err error
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantFound bool
	}{
		{
			name:   "without err",
			fields: fields{},
			args: args{
				"foo",
				nil,
				nil,
			},
			wantErr:   false,
			wantFound: true,
		},
		{
			name:   "with err",
			fields: fields{},
			args: args{
				"foo",
				nil,
				errors.New("bar"),
			},
			wantErr:   true,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ff := &Finder{
				SearchDirectories: tt.fields.SearchDirectories,
				foundFiles:        tt.fields.foundFiles,
			}
			if err := ff.markSeen(tt.args.in0, tt.args.in1, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Finder.markSeen() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFound != ff.foundFiles {
				t.Errorf("Finder.markSeen() foundFiles = %v, wantFound %v", ff.foundFiles, tt.wantFound)
			}
		})
	}
}

func TestFinder_PrependSearchDirectory(t *testing.T) {
	t.Parallel()
	type fields struct {
		SearchDirectories []string
		foundFiles        bool
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantDirs []string
	}{
		{
			name: "without trailing /",
			fields: fields{
				SearchDirectories: []string{"../testdata/foo/"},
			},
			args: args{
				dir: "../testdata/bar",
			},
			wantDirs: []string{
				"../testdata/bar/",
				"../testdata/foo/",
			},
		},
		{
			name: "with trailing /",
			fields: fields{
				SearchDirectories: []string{"../testdata/foo/"},
			},
			args: args{
				dir: "../testdata/bar/",
			},
			wantDirs: []string{
				"../testdata/bar/",
				"../testdata/foo/",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ff := &Finder{
				SearchDirectories: tt.fields.SearchDirectories,
				foundFiles:        tt.fields.foundFiles,
			}
			ff.PrependSearchDirectory(tt.args.dir)
			if !reflect.DeepEqual(ff.SearchDirectories, tt.wantDirs) {
				t.Errorf("Finder.SearchDirectories = %v, want %v", ff.SearchDirectories, tt.wantDirs)
			}
		})
	}
}

func TestFinder_AppendSearchDirectory(t *testing.T) {
	t.Parallel()
	type fields struct {
		SearchDirectories []string
		foundFiles        bool
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantDirs []string
	}{
		{
			name: "without trailing /",
			fields: fields{
				SearchDirectories: []string{"../testdata/foo/"},
			},
			args: args{
				dir: "../testdata/bar",
			},
			wantDirs: []string{
				"../testdata/foo/",
				"../testdata/bar/",
			},
		},
		{
			name: "with trailing /",
			fields: fields{
				SearchDirectories: []string{"../testdata/foo/"},
			},
			args: args{
				dir: "../testdata/bar/",
			},
			wantDirs: []string{
				"../testdata/foo/",
				"../testdata/bar/",
			},
		},
		{
			name: "findall",
			fields: fields{
				SearchDirectories: []string{},
			},
			args: args{
				dir: "../testdata/findall",
			},
			wantDirs: []string{
				"../testdata/findall/",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ff := &Finder{
				SearchDirectories: tt.fields.SearchDirectories,
				foundFiles:        tt.fields.foundFiles,
			}
			ff.AppendSearchDirectory(tt.args.dir)
			if !reflect.DeepEqual(ff.SearchDirectories, tt.wantDirs) {
				t.Errorf("Finder.SearchDirectories = %v, want %v", ff.SearchDirectories, tt.wantDirs)
			}
		})
	}
}

func TestFinder_Find(t *testing.T) {
	t.Parallel()
	type fields struct {
		SearchDirectories []string
		foundFiles        bool
	}
	type args struct {
		filePat string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantPattern *pattern.Pattern
		wantErr     bool
	}{
		{
			name: "testdata/foo/bar found by default dirs",
			fields: fields{
				SearchDirectories: []string{"../testdata/foo/"},
			},
			args:        args{"bar"},
			wantPattern: pattern.NewPatternPtr("../testdata/foo/bar"),
			wantErr:     false,
		},
		{
			name: "testdata/foo/bar found by explicit pattern",
			fields: fields{
				SearchDirectories: []string{"./"},
			},
			args:        args{"../testdata/foo/bar"},
			wantPattern: pattern.NewPatternPtr("../testdata/foo/bar"),
			wantErr:     false,
		},
		{
			name: "bar not found",
			fields: fields{
				SearchDirectories: []string{},
			},
			args:        args{"bar"},
			wantPattern: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ff := &Finder{
				SearchDirectories: tt.fields.SearchDirectories,
				foundFiles:        tt.fields.foundFiles,
			}
			gotPattern, err := ff.Find(context.Background(), tt.args.filePat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Finder.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !pattern.FilePatternsEqual(gotPattern, tt.wantPattern) {
				t.Errorf("Finder.Find() = %v, want %v", gotPattern, tt.wantPattern)
			}
		})
	}
}

func Test_cleanDirName(t *testing.T) {
	t.Parallel()
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with trailing /",
			args: args{
				"foo",
			},
			want: "foo/",
		},
		{
			name: "without trailing /",
			args: args{
				"foo/",
			},
			want: "foo/",
		},
		{
			name: "findall",
			args: args{
				"../testdata/findall",
			},
			want: "../testdata/findall/",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := cleanDirName(tt.args.dir); got != tt.want {
				t.Errorf("cleanDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dirExists(t *testing.T) {
	t.Parallel()
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "existing dir",
			args: args{dir: "../testdata/bar"},
			want: true,
		},
		{
			name: "non-existing dir",
			args: args{dir: "../testdata/notthere"},
			want: false,
		},
		{
			name: "existing non-dir",
			args: args{dir: "../testdata/bar/baz-1.jpl"},
			want: false,
		},
		{
			name: "findall",
			args: args{dir: "../testdata/findall"},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := dirExists(tt.args.dir); got != tt.want {
				t.Errorf("dirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinder_FindAllFromPattern(t *testing.T) {
	t.Parallel()
	type args struct {
		filePat string
	}
	tests := []struct {
		name         string
		ff           *Finder
		args         args
		wantPatterns []*pattern.Pattern
		wantErr      bool
	}{
		{
			name: "just testdata",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/"},
			},
			args:         args{filePat: "bar"},
			wantPatterns: []*pattern.Pattern{},
			wantErr:      false,
		},
		{
			name: "testdata all subdirs multiple",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/bar/", "../testdata/findall/", "../testdata/foo/", "../testdata/taccat/"},
			},
			args: args{filePat: "bar"},
			wantPatterns: []*pattern.Pattern{
				pattern.NewPatternPtr("../testdata/findall/bar"),
				pattern.NewPatternPtr("../testdata/foo/bar"),
			},
			wantErr: false,
		},
		{
			name: "testdata all subdirs single",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/bar/", "../testdata/findall/", "../testdata/foo/", "../testdata/taccat/"},
			},
			args: args{filePat: "baz"},
			wantPatterns: []*pattern.Pattern{
				pattern.NewPatternPtr("../testdata/bar/baz"),
			},
			wantErr: false,
		},
		{
			name: "testdata some subdirs",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/bar/", "../testdata/foo/", "../testdata/taccat/"},
			},
			args: args{filePat: "bar"},
			wantPatterns: []*pattern.Pattern{
				pattern.NewPatternPtr("../testdata/foo/bar"),
			},
			wantErr: false,
		},
		{
			name: "empty dirname",
			ff: &Finder{
				SearchDirectories: []string{},
			},
			args: args{filePat: "foo/bar"},
			wantPatterns: []*pattern.Pattern{
				pattern.NewPatternPtrWithOriginal("foo/bar", "foo/bar"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPatterns, err := tt.ff.FindAllFromPattern(context.Background(), tt.args.filePat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Finder.FindAllFromPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gotPatterns) != len(tt.wantPatterns) {
				t.Errorf("Finder.FindAllFromPattern() len = %v, want %v", gotPatterns, tt.wantPatterns)
				return
			}

			for i, fp1 := range gotPatterns {
				i, fp1 := i, fp1
				fp2 := tt.wantPatterns[i]
				if !pattern.FilePatternsEqual(fp1, fp2) {
					t.Errorf("Finder.FindAllFromPattern()[%d] = %v, want %v", i, *fp1, *fp2)
				}
			}
		})
	}
}

func TestFinder_FindAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		ff           *Finder
		wantPatterns []*pattern.Pattern
		wantErr      bool
	}{
		{
			name: "Find all in testdata",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/"},
			},
			wantPatterns: []*pattern.Pattern{},
			wantErr:      false,
		},
		{
			name: "Find all in testdata subdirs",
			ff: &Finder{
				SearchDirectories: []string{"../testdata/bar/", "../testdata/findall/", "../testdata/foo/", "../testdata/taccat/", "../testdata/find/"},
			},
			wantPatterns: []*pattern.Pattern{
				pattern.NewPatternPtr("../testdata/bar/baz.jpl"),
				pattern.NewPatternPtr("../testdata/findall/bar.jpl"),
				pattern.NewPatternPtr("../testdata/findall/foo.jpl"),
				pattern.NewPatternPtr("../testdata/foo/bar.log"),
				pattern.NewPatternPtr("../testdata/taccat/test.log"),
				pattern.NewPatternPtr("../testdata/find/test-tsar.thing.log"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPatterns, err := tt.ff.FindAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Finder.FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, fp := range gotPatterns {
				fmt.Println(fp.Pattern())
			}

			if len(gotPatterns) != len(tt.wantPatterns) {
				t.Errorf("Finder.FindAll() len = %v, want %v", len(gotPatterns), len(tt.wantPatterns))
				return
			}

			for i, fp1 := range gotPatterns {
				fp2 := tt.wantPatterns[i]
				if !pattern.FilePatternsEqual(fp1, fp2) {
					t.Errorf("Finder.FindAll()[%d] = %v, want %v", i, *fp1, *fp2)
				}
			}
		})
	}
}
