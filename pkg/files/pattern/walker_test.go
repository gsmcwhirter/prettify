package pattern

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/pathutil"
)

type walkRecorderRecord struct {
	path string
	err  error
}

type walkRecorder struct {
	Records []walkRecorderRecord
}

func (wr *walkRecorder) Record(path string, info os.FileInfo, err error) error {
	wr.Records = append(wr.Records, walkRecorderRecord{
		path: path,
		err:  err,
	})
	return nil
}

func Test_walker_walkFilter(t *testing.T) {
	t.Parallel()
	type args struct {
		path string
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name        string
		recorder    walkRecorder
		fpw         *walker
		args        args
		wantRecords []walkRecorderRecord
		wantErr     bool
	}{
		{
			name:     "match no filters",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "foo",
					directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
					filePattern:   "foo",
					filenameGlob:  formFileNameGlob("foo", ""),
					filenameRegex: formFileNameRegexp("foo", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			args: args{
				path: pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl",
				info: nil,
				err:  nil,
			},
			wantRecords: []walkRecorderRecord{{pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl", nil}},
			wantErr:     false,
		},
		{
			name:     "fail fast",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "foo",
					directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
					filePattern:   "foo",
					filenameGlob:  formFileNameGlob("foo", ""),
					filenameRegex: formFileNameRegexp("foo", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			args: args{
				path: pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl",
				info: nil,
				err:  errors.New("test"),
			},
			wantRecords: []walkRecorderRecord{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fpw.setWalkFunc(tt.recorder.Record)
			if err := tt.fpw.walkFilter(context.Background())(tt.args.path, tt.args.info, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("walker.walkFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.recorder.Records, tt.wantRecords) && (len(tt.recorder.Records) > 0 || len(tt.wantRecords) > 0) {
				t.Errorf("walker.walkFilter() records = %v, want %v", tt.recorder.Records, tt.wantRecords)
			}
		})
	}
}

func Test_walker_walk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		recorder    walkRecorder
		fpw         *walker
		wantRecords []walkRecorderRecord
		wantErr     bool
	}{
		{
			name:     "match one no filters",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "foo",
					directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
					filePattern:   "foo",
					filenameGlob:  formFileNameGlob("foo", ""),
					filenameRegex: formFileNameRegexp("foo", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl", nil},
			},
			wantErr: false,
		},
		{
			name:     "match several no filters",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "bar",
					directory:     pathutil.MustAbsPath("../testdata/foo") + "/",
					filePattern:   "bar",
					filenameGlob:  formFileNameGlob("bar", ""),
					filenameRegex: formFileNameRegexp("bar", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
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
			tt.fpw.setWalkFunc(tt.recorder.Record)
			if err := tt.fpw.walk(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("walker.walk() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.recorder.Records, tt.wantRecords) && (len(tt.recorder.Records) > 0 || len(tt.wantRecords) > 0) {
				t.Errorf("walker.walk() records = %v, want %v", tt.recorder.Records, tt.wantRecords)
			}
		})
	}
}

func Test_walker_setWalkFunc(t *testing.T) {
	t.Parallel()
	type args struct {
		walkFunc filepath.WalkFunc
	}
	tests := []struct {
		name string
		fpw  *walker
		args args
	}{
		{
			name: "basic test",
			fpw: &walker{
				filePattern: nil,
				walkFunc:    nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := walkRecorder{[]walkRecorderRecord{}}
			tt.args = args{recorder.Record}
			tt.fpw.setWalkFunc(tt.args.walkFunc)

			if tt.fpw.walkFunc == nil {
				t.Errorf("walker.setWalkFunc() set %p, want %p", tt.fpw.walkFunc, recorder.Record)
			}
		})
	}
}

func Test_walker_walkReverse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		recorder    walkRecorder
		fpw         *walker
		wantRecords []walkRecorderRecord
		wantErr     bool
	}{
		{
			name:     "match one no filters",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "foo",
					directory:     pathutil.MustAbsPath("../testdata/findall") + "/",
					filePattern:   "foo",
					filenameGlob:  formFileNameGlob("foo", ""),
					filenameRegex: formFileNameRegexp("foo", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/findall") + "/foo-out-1.jpl", nil},
			},
			wantErr: false,
		},
		{
			name:     "match several no filters",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "bar",
					directory:     pathutil.MustAbsPath("../testdata/foo") + "/",
					filePattern:   "bar",
					filenameGlob:  formFileNameGlob("bar", ""),
					filenameRegex: formFileNameRegexp("bar", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-out.log", nil},
				{pathutil.MustAbsPath("../testdata/foo") + "/bar-error.log", nil},
			},
			wantErr: false,
		},
		{
			name:     "lexical order",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "test",
					directory:     pathutil.MustAbsPath("../testdata/zeros") + "/",
					filePattern:   "test",
					filenameGlob:  formFileNameGlob("test", ""),
					filenameRegex: formFileNameRegexp("test", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/zeros") + "/test-out-212.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros") + "/test-out-201.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros") + "/test-out-111.log", nil},
			},
			wantErr: false,
		},
		{
			name:     "lexical order 2",
			recorder: walkRecorder{[]walkRecorderRecord{}},
			fpw: &walker{
				filePattern: &Pattern{
					origPattern:   "job_alerts_sender",
					directory:     pathutil.MustAbsPath("../testdata/zeros2") + "/",
					filePattern:   "foobar",
					filenameGlob:  formFileNameGlob("foober", ""),
					filenameRegex: formFileNameRegexp("foobar", ""),
				},
				walkFunc: nil, // set right before running the test for "reasons"
			},
			wantRecords: []walkRecorderRecord{
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-out-4.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-out-3.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-out-2.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-out-1.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-error-3.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-error-2.log", nil},
				{pathutil.MustAbsPath("../testdata/zeros2") + "/foobar-error-1.log", nil},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fpw.setWalkFunc(tt.recorder.Record)
			if err := tt.fpw.walkReverse(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("walker.walkReverse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.recorder.Records, tt.wantRecords) && (len(tt.recorder.Records) > 0 || len(tt.wantRecords) > 0) {
				t.Errorf("walker.walkReverse() records = %v, want %v", tt.recorder.Records, tt.wantRecords)
			}
		})
	}
}
