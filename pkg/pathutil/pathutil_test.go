package pathutil

import "testing"

// This basically just tests to make sure stuff compiles cleanly

func TestMustAbsPath(t *testing.T) {
	t.Parallel()
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "works",
			args: args{
				path: "test",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = MustAbsPath(tt.args.path)
		})
	}
}
