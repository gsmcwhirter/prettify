package watcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/go-util/v9/deferutil"

	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
	"github.com/gsmcwhirter/prettify/pkg/pathutil"
)

func TestNewWatcher(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	type args struct {
		fp *pattern.Pattern
	}
	tests := []struct {
		name string
		args args
		want *Watcher
	}{
		{
			name: "basic test",
			args: args{
				fp: &testFp,
			},
			want: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   make([]string, 16),
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewWatcher(tt.args.fp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatcher_resetSeenThisTime(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	tests := []struct {
		name              string
		fw                *Watcher
		wantSeenPreSlice  []string
		wantSeenPostSlice []string
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles: map[string]bool{
					"foo": true,
					"bar": true,
					"baz": true,
				},
				seenThisTime: []string{
					"foo", "bar", "baz",
				},
				seenThisTimeCt: 2,
				filePattern:    &testFp,
			},
			wantSeenPreSlice:  []string{"foo", "bar"},
			wantSeenPostSlice: []string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fw.SeenThisTime(); !reflect.DeepEqual(got, tt.wantSeenPreSlice) {
				t.Errorf("pre-reset seenThisTime = %v, want %v", got, tt.wantSeenPreSlice)
			}

			tt.fw.resetSeenThisTime()

			if got := tt.fw.SeenThisTime(); !reflect.DeepEqual(got, tt.wantSeenPostSlice) {
				t.Errorf("post-reset seenThisTime = %v, want %v", got, tt.wantSeenPostSlice)
			}
		})
	}
}

func TestWatcher_seeFile(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	type args struct {
		path string
		in1  os.FileInfo
		err  error
	}
	tests := []struct {
		name       string
		fw         *Watcher
		args       args
		wantMap    map[string]bool
		wantSeen   []string
		wantSeenCt int
		wantErr    bool
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   make([]string, 16),
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
			args: args{
				path: "foo",
				in1:  nil,
				err:  nil,
			},
			wantMap: map[string]bool{
				"foo": true,
			},
			wantSeen:   []string{"foo", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
			wantSeenCt: 1,
			wantErr:    false,
		},
		{
			name: "error test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   make([]string, 16),
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
			args: args{
				path: "foo",
				in1:  nil,
				err:  errors.New("test"),
			},
			wantMap:    map[string]bool{},
			wantSeen:   make([]string, 16),
			wantSeenCt: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.fw.seeFile(tt.args.path, tt.args.in1, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Watcher.seeFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := tt.fw.SeenFiles; !reflect.DeepEqual(got, tt.wantMap) {
				t.Errorf("SeenFiles = %v, want %v", got, tt.wantMap)
			}

			if got := tt.fw.seenThisTime; !reflect.DeepEqual(got, tt.wantSeen) {
				t.Errorf("seenThisTime = %v, want %v", got, tt.wantSeen)
			}

			if got := tt.fw.seenThisTimeCt; !reflect.DeepEqual(got, tt.wantSeenCt) {
				t.Errorf("seenThisTimeCt = %v, want %v", got, tt.wantSeenCt)
			}
		})
	}
}

func TestWatcher_addToSeenThisTime(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	type args struct {
		path string
	}
	tests := []struct {
		name       string
		fw         *Watcher
		args       args
		wantSeen   []string
		wantSeenCt int
	}{
		{
			name: "resize test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   []string{"foo", "foo", "foo", "foo"},
				seenThisTimeCt: 4,
				filePattern:    &testFp,
			},
			args: args{
				path: "bar",
			},
			wantSeen:   []string{"foo", "foo", "foo", "foo", "bar", "", "", ""},
			wantSeenCt: 5,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fw.addToSeenThisTime(tt.args.path)

			if got := tt.fw.seenThisTime; !reflect.DeepEqual(got, tt.wantSeen) {
				t.Errorf("seenThisTime = %v, want %v", got, tt.wantSeen)
			}

			if got := tt.fw.seenThisTimeCt; !reflect.DeepEqual(got, tt.wantSeenCt) {
				t.Errorf("seenThisTimeCt = %v, want %v", got, tt.wantSeenCt)
			}
		})
	}
}

func TestWatcher_LastSeenThisTime(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	tests := []struct {
		name    string
		fw      *Watcher
		want    string
		wantErr bool
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   []string{"foo", "bar", "baz", "quux"},
				seenThisTimeCt: 2,
				filePattern:    &testFp,
			},
			want:    "bar",
			wantErr: false,
		},
		{
			name: "error test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   []string{"foo", "bar", "baz", "quux"},
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.fw.LastSeenThisTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("Watcher.LastSeenThisTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Watcher.LastSeenThisTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatcher_SeenThisTime(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	tests := []struct {
		name string
		fw   *Watcher
		want []string
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   []string{"foo", "bar", "baz", "quux"},
				seenThisTimeCt: 3,
				filePattern:    &testFp,
			},
			want: []string{"foo", "bar", "baz"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fw.SeenThisTime(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Watcher.SeenThisTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatcher_Run(t *testing.T) {
	t.Parallel()
	testFp := pattern.NewPattern(pathutil.MustAbsPath("../testdata/taccat/test"))

	tests := []struct {
		name     string
		fw       *Watcher
		wantSeen []string
		wantErr  bool
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   make([]string, 16),
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
			wantSeen: []string{
				pathutil.MustAbsPath("../testdata/taccat/test-out-1.log"),
				pathutil.MustAbsPath("../testdata/taccat/test-out-11.log"),
				pathutil.MustAbsPath("../testdata/taccat/test-out-2.log"),
				pathutil.MustAbsPath("../testdata/taccat/test-out-3.log"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.fw.Run(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Watcher.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := tt.fw.SeenThisTime(); !reflect.DeepEqual(got, tt.wantSeen) {
				t.Errorf("Watcher.SeenThisTime() = %v, want %v", got, tt.wantSeen)
			}
		})
	}
}

func TestWatcher_Run2(t *testing.T) {
	t.Parallel()
	testDir, err := ioutil.TempDir(os.TempDir(), "TestWatcher_Run2")
	if err != nil {
		t.Error("Could not create temporary test directory")
		return
	}
	t.Cleanup(func() {
		err := os.RemoveAll(testDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not remove directory %s: %s", testDir, err)
		}
	})

	sourceDir := pathutil.MustAbsPath("../testdata/taccat")
	mustCopyFile := func(name string) {
		from, err := os.Open(sourceDir + "/" + name)
		if err != nil {
			panic(err)
		}
		defer deferutil.CheckDefer(from.Close)

		to, err := os.Create(testDir + "/" + name)
		if err != nil {
			panic(err)
		}
		defer deferutil.CheckDefer(to.Close)

		_, err = io.Copy(to, from)
		if err != nil {
			panic(err)
		}
	}

	mustCopyFile("test-out-1.log")
	mustCopyFile("test-out-11.log")
	mustCopyFile("test-out-2.log")
	mustCopyFile("test-out-3.log")

	testFp := pattern.NewPattern(pathutil.MustAbsPath(testDir + "/test"))

	tests := []struct {
		name             string
		fw               *Watcher
		wantSeenLastTime [][]string
		wantErr          bool
	}{
		{
			name: "basic test",
			fw: &Watcher{
				SeenFiles:      map[string]bool{},
				seenThisTime:   make([]string, 16),
				seenThisTimeCt: 0,
				filePattern:    &testFp,
			},
			wantSeenLastTime: [][]string{
				{
					pathutil.MustAbsPath(testDir + "/test-out-1.log"),
					pathutil.MustAbsPath(testDir + "/test-out-11.log"),
					pathutil.MustAbsPath(testDir + "/test-out-2.log"),
					pathutil.MustAbsPath(testDir + "/test-out-3.log"),
				},
				{
					pathutil.MustAbsPath(testDir + "/test-out-3.log"),
					pathutil.MustAbsPath(testDir + "/test-out-4.log"),
				},
				{
					pathutil.MustAbsPath(testDir + "/test-out-4.log"),
					pathutil.MustAbsPath(testDir + "/test-out-5.log"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i, wslt := range tt.wantSeenLastTime {
				i, wslt := i, wslt
				if err := tt.fw.Run(context.Background()); (err != nil) != tt.wantErr {
					t.Errorf("Watcher.Run() run %d error = %v, wantErr %v", i, err, tt.wantErr)
				}

				if got := tt.fw.SeenThisTime(); !reflect.DeepEqual(got, wslt) {
					t.Errorf("Watcher.SeenThisTime() run %d = %v, want %v", i, got, wslt)
				}

				fname := fmt.Sprintf(pathutil.MustAbsPath(testDir+"/test-out-%d.log"), 4+i)
				fh, err := os.Create(fname)
				if err != nil {
					t.Errorf("Watcher.SeenThisTime() post run %d could not create %s", i, fname)
				} else {
					fh.Close()
					defer os.Remove(fname)
				}
			}
		})
	}
}
