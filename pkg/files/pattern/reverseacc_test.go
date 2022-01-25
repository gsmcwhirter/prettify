package pattern

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func Test_reverseAccumulator_Record(t *testing.T) {
	t.Parallel()
	type args struct {
		path string
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name      string
		ra        *reverseAccumulator
		args      args
		wantPaths []reverseAccumulatorRecord
		wantErr   bool
	}{
		{
			name: "record without error",
			ra: &reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "bar", info: nil},
				},
			},
			args: args{
				path: "foo",
				info: nil,
				err:  nil,
			},
			wantPaths: []reverseAccumulatorRecord{
				{path: "bar", info: nil},
				{path: "foo", info: nil},
			},
			wantErr: false,
		},
		{
			name: "record with error",
			ra: &reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "bar", info: nil},
				},
			},
			args: args{
				path: "foo",
				info: nil,
				err:  errors.New("test"),
			},
			wantPaths: []reverseAccumulatorRecord{
				{path: "bar", info: nil},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.ra.record(tt.args.path, tt.args.info, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("reverseAccumulator.Record() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.ra.Paths, tt.wantPaths) && (len(tt.ra.Paths) > 0 || len(tt.wantPaths) > 0) {
				t.Errorf("reverseAccumulator.Record() = %v, want %v", tt.ra.Paths, tt.wantPaths)
			}
		})
	}
}

func Test_reverseAccumulator_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ra   reverseAccumulator
		want int
	}{
		{
			name: "empty",
			ra:   reverseAccumulator{},
			want: 0,
		},
		{
			name: "non-empty",
			ra: reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "foo", info: nil},
					{path: "bar", info: nil},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ra.Len(); got != tt.want {
				t.Errorf("reverseAccumulator.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseAccumulator_Swap(t *testing.T) {
	t.Parallel()
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name      string
		ra        reverseAccumulator
		args      args
		wantPaths []reverseAccumulatorRecord
	}{
		{
			name: "swap 0,1",
			ra: reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "foo", info: nil},
					{path: "bar", info: nil},
					{path: "baz", info: nil},
				},
			},
			args: args{i: 0, j: 1},
			wantPaths: []reverseAccumulatorRecord{
				{path: "bar", info: nil},
				{path: "foo", info: nil},
				{path: "baz", info: nil},
			},
		},
		{
			name: "swap 0,2",
			ra: reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "foo", info: nil},
					{path: "bar", info: nil},
					{path: "baz", info: nil},
				},
			},
			args: args{i: 0, j: 2},
			wantPaths: []reverseAccumulatorRecord{
				{path: "baz", info: nil},
				{path: "bar", info: nil},
				{path: "foo", info: nil},
			},
		},
		{
			name: "swap 1,2",
			ra: reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "foo", info: nil},
					{path: "bar", info: nil},
					{path: "baz", info: nil},
				},
			},
			args: args{i: 1, j: 2},
			wantPaths: []reverseAccumulatorRecord{
				{path: "foo", info: nil},
				{path: "baz", info: nil},
				{path: "bar", info: nil},
			},
		},
		{
			name: "swap 1,1",
			ra: reverseAccumulator{
				Paths: []reverseAccumulatorRecord{
					{path: "foo", info: nil},
					{path: "bar", info: nil},
					{path: "baz", info: nil},
				},
			},
			args: args{i: 1, j: 1},
			wantPaths: []reverseAccumulatorRecord{
				{path: "foo", info: nil},
				{path: "bar", info: nil},
				{path: "baz", info: nil},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.ra.Swap(tt.args.i, tt.args.j)

			if !reflect.DeepEqual(tt.ra.Paths, tt.wantPaths) {
				t.Errorf("reverseAccumulator.Swap(%d, %d) = %v, want %v", tt.args.i, tt.args.j, tt.ra.Paths, tt.wantPaths)
			}
		})
	}
}

func Test_reverseAccumulator_Less(t *testing.T) {
	t.Parallel()
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		ra   reverseAccumulator
		args args
		want bool
	}{
		{
			name: "same index",
			ra:   reverseAccumulator{},
			args: args{i: 0, j: 0},
			want: false,
		},
		{
			name: "i < j",
			ra:   reverseAccumulator{},
			args: args{i: 0, j: 1},
			want: false,
		},
		{
			name: "i > j",
			ra:   reverseAccumulator{},
			args: args{i: 1, j: 0},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ra.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("reverseAccumulator.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}
