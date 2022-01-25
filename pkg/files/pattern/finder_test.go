package pattern

import (
	"context"
	"errors"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestNewFinder(t *testing.T) {
	t.Parallel()
	type args struct {
		directory string
	}
	tests := []struct {
		name string
		args args
		want Finder
	}{
		{
			name: "basic test",
			args: args{"../testdata/"},
			want: Finder{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewFinder(tt.args.directory); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFinder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinder_SeenPatterns(t *testing.T) {
	t.Parallel()
	type fields struct {
		directory    string
		seenPatterns map[string]bool
		matchRegex   *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "empty",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			want: []string{},
		},
		{
			name: "non empty",
			fields: fields{
				directory: "../testdata/",
				seenPatterns: map[string]bool{
					"test2": true,
					"test":  true,
				},
				matchRegex: patternFilenameRegex,
			},
			want: []string{"test", "test2"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pf := &Finder{
				directory:    tt.fields.directory,
				seenPatterns: tt.fields.seenPatterns,
				matchRegex:   tt.fields.matchRegex,
			}
			if got := pf.SeenPatterns(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Finder.SeenPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinder_walkFilter(t *testing.T) {
	t.Parallel()
	type fields struct {
		directory    string
		seenPatterns map[string]bool
		matchRegex   *regexp.Regexp
	}
	type args struct {
		path string
		in1  os.FileInfo
		err  error
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSeen map[string]bool
		wantErr  bool
	}{
		{
			name: "does match",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			args: args{
				path: "../testdata/test-out-1.log",
				in1:  nil,
				err:  nil,
			},
			wantSeen: map[string]bool{"../testdata/test.log": true},
			wantErr:  false,
		},
		{
			name: "no match",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			args: args{
				path: "../testdata/test.log",
				in1:  nil,
				err:  nil,
			},
			wantSeen: map[string]bool{},
			wantErr:  false,
		},
		{
			name: "fast error",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			args: args{
				path: "../testdata/test-1.log",
				in1:  nil,
				err:  errors.New("test"),
			},
			wantSeen: map[string]bool{},
			wantErr:  true,
		},
		{
			name: "no regexp error",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   nil,
			},
			args: args{
				path: "../testdata/test.log",
				in1:  nil,
				err:  nil,
			},
			wantSeen: map[string]bool{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pf := &Finder{
				directory:    tt.fields.directory,
				seenPatterns: tt.fields.seenPatterns,
				matchRegex:   tt.fields.matchRegex,
			}
			if err := pf.walkFilter(context.Background())(tt.args.path, tt.args.in1, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Finder.walkFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(pf.seenPatterns, tt.wantSeen) {
				t.Errorf("Finder.walkFilter() seen = %v, want %v", pf.seenPatterns, tt.wantSeen)
			}
		})
	}
}

func TestFinder_Walk(t *testing.T) {
	t.Parallel()
	type fields struct {
		directory    string
		seenPatterns map[string]bool
		matchRegex   *regexp.Regexp
	}
	tests := []struct {
		name     string
		fields   fields
		wantSeen map[string]bool
		wantErr  bool
	}{
		{
			name: "testdata",
			fields: fields{
				directory:    "../testdata/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			wantSeen: map[string]bool{},
			wantErr:  false,
		},
		{
			name: "testdata/taccat",
			fields: fields{
				directory:    "../testdata/taccat/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			wantSeen: map[string]bool{
				"../testdata/taccat/test.log": true,
			},
			wantErr: false,
		},
		{
			name: "testdata/findall",
			fields: fields{
				directory:    "../testdata/findall/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			wantSeen: map[string]bool{
				"../testdata/findall/foo.jpl": true,
				"../testdata/findall/bar.jpl": true,
			},
			wantErr: false,
		},
		{
			name: "testdata/find",
			fields: fields{
				directory:    "../testdata/find/",
				seenPatterns: map[string]bool{},
				matchRegex:   patternFilenameRegex,
			},
			wantSeen: map[string]bool{
				"../testdata/find/test-tsar.thing.log": true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pf := &Finder{
				directory:    tt.fields.directory,
				seenPatterns: tt.fields.seenPatterns,
				matchRegex:   tt.fields.matchRegex,
			}
			if err := pf.Walk(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Finder.Walk() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(pf.seenPatterns, tt.wantSeen) {
				t.Errorf("Finder.Walk() seen = %v, want %v", pf.seenPatterns, tt.wantSeen)
			}
		})
	}
}
