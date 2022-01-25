package tagfinder

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/gsmcwhirter/prettify/pkg/pathutil"
)

var testDataDir = "../../pkg/files/testdata"

func TestNewTagFinder(t *testing.T) {
	t.Parallel()
	type args struct {
		sampleSize uint
		findAll    bool
	}
	tests := []struct {
		name string
		args args
		want TagFinder
	}{
		{
			name: "basic test",
			args: args{
				sampleSize: 10,
				findAll:    false,
			},
			want: TagFinder{
				tags:       map[string]bool{},
				sampleSize: 10,
				findAll:    false,
			},
		},
		{
			name: "basic test 2",
			args: args{
				sampleSize: 100,
				findAll:    true,
			},
			want: TagFinder{
				tags:       map[string]bool{},
				sampleSize: 100,
				findAll:    true,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewTagFinder(tt.args.sampleSize, tt.args.findAll); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTagFinder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagFinder_Tags(t *testing.T) {
	t.Parallel()
	type fields struct {
		tags       map[string]bool
		sampleSize uint
		findAll    bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "basic test",
			fields: fields{
				tags: map[string]bool{
					"foo": true,
					"bar": true,
					"baz": true,
				},
				sampleSize: 10,
				findAll:    false,
			},
			want: []string{"bar", "baz", "foo"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := &TagFinder{
				tags:       tt.fields.tags,
				sampleSize: tt.fields.sampleSize,
				findAll:    tt.fields.findAll,
			}
			if got := tf.Tags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TagFinder.Tags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractTag(t *testing.T) {
	t.Parallel()
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "valid json with @tag",
			args:    args{`{"@tag": "foo", "not_tag": 1}`},
			want:    "foo",
			wantErr: false,
		},
		{
			name:    "valid json no @tag",
			args:    args{`{"not_tag": 1}`},
			want:    "",
			wantErr: true,
		},
		{
			name:    "in-valid json-like",
			args:    args{`{@tag is not here}`},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := extractTag(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagFinder_Walker(t *testing.T) {
	t.Parallel()
	type fields struct {
		tags       map[string]bool
		sampleSize uint
		findAll    bool
	}
	type args struct {
		path string
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantTags []string
		wantErr  bool
	}{
		{
			name: "basic test long sample",
			fields: fields{
				tags:       map[string]bool{},
				sampleSize: 10,
				findAll:    false,
			},
			args: args{
				path: pathutil.MustAbsPath(fmt.Sprintf("%s/tagfind/test-out-1.log", testDataDir)),
				info: nil,
				err:  nil,
			},
			wantTags: []string{"test1"},
			wantErr:  false,
		},
		{
			name: "basic test short sample",
			fields: fields{
				tags:       map[string]bool{},
				sampleSize: 1,
				findAll:    false,
			},
			args: args{
				path: pathutil.MustAbsPath(fmt.Sprintf("%s/tagfind/test-out-2.log", testDataDir)),
				info: nil,
				err:  nil,
			},
			wantTags: []string{"test1"},
			wantErr:  false,
		},
		{
			name: "basic test short sample findAll",
			fields: fields{
				tags:       map[string]bool{},
				sampleSize: 1,
				findAll:    true,
			},
			args: args{
				path: pathutil.MustAbsPath(fmt.Sprintf("%s/tagfind/test-out-2.log", testDataDir)),
				info: nil,
				err:  nil,
			},
			wantTags: []string{"test1", "test2"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := &TagFinder{
				tags:       tt.fields.tags,
				sampleSize: tt.fields.sampleSize,
				findAll:    tt.fields.findAll,
			}
			if err := tf.Walker(tt.args.path, tt.args.info, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("TagFinder.Walker() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := tf.Tags(); !reflect.DeepEqual(got, tt.wantTags) {
				t.Errorf("TagFinder.Walker() tags = %v, wantTags %v", got, tt.wantTags)
			}
		})
	}
}
