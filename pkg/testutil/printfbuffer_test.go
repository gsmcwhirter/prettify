package testutil

import (
	"reflect"
	"testing"
)

func TestNewPrintfBuffer(t *testing.T) {
	t.Parallel()
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want PrintfBuffer
	}{
		{
			name: "test basic",
			args: args{
				size: 128,
			},
			want: PrintfBuffer{
				writtenData:  make([]byte, 128),
				bytesWritten: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewPrintfBuffer(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrintfBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintfBuffer_expandIfNecessary(t *testing.T) {
	t.Parallel()
	type pbData struct {
		writtenData  []byte
		bytesWritten int
	}
	type args struct {
		addingBytes int
	}
	tests := []struct {
		name     string
		pbData   pbData
		args     args
		wantSize int
	}{
		{
			name: "no expand",
			pbData: pbData{
				writtenData:  make([]byte, 16),
				bytesWritten: 0,
			},
			args: args{
				addingBytes: 16,
			},
			wantSize: 16,
		},
		{
			name: "expand",
			pbData: pbData{
				writtenData:  []byte("abcdefghijklmnop"),
				bytesWritten: 12,
			},
			args: args{
				addingBytes: 16,
			},
			wantSize: 32,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := PrintfBuffer{
				writtenData:  tt.pbData.writtenData,
				bytesWritten: tt.pbData.bytesWritten,
			}
			pb.expandIfNecessary(tt.args.addingBytes)
			if len(pb.writtenData) != tt.wantSize {
				t.Errorf("expandIfNecessary expected new size %d got %d", tt.wantSize, len(pb.writtenData))
			}
			if !reflect.DeepEqual(pb.writtenData[:tt.pbData.bytesWritten], tt.pbData.writtenData[:tt.pbData.bytesWritten]) {
				t.Errorf("expandIfNecessary mangled existing data -- got %v expected %v", pb.writtenData[:tt.pbData.bytesWritten], tt.pbData.writtenData[:tt.pbData.bytesWritten])
			}
		})
	}
}

func TestPrintfBuffer_writeBytes(t *testing.T) {
	t.Parallel()
	type pbData struct {
		writtenData  []byte
		bytesWritten int
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name        string
		pbData      pbData
		args        args
		wantWritten int
		wantBytes   []byte
	}{
		{
			name: "no expand",
			pbData: pbData{
				writtenData:  make([]byte, 16),
				bytesWritten: 0,
			},
			args: args{
				data: []byte("test"),
			},
			wantWritten: 4,
			wantBytes:   []byte("test"),
		},
		{
			name: "expand",
			pbData: pbData{
				writtenData:  []byte("abcdefghijklmnop"),
				bytesWritten: 12,
			},
			args: args{
				data: []byte("testtest"),
			},
			wantWritten: 20,
			wantBytes:   []byte("abcdefghijkltesttest"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := PrintfBuffer{
				writtenData:  tt.pbData.writtenData,
				bytesWritten: tt.pbData.bytesWritten,
			}
			pb.writeBytes(tt.args.data)
			if pb.bytesWritten != tt.wantWritten {
				t.Errorf("writeData expected bytesWritten %d got %d", tt.wantWritten, pb.bytesWritten)
			}
			if !reflect.DeepEqual(pb.writtenData[:pb.bytesWritten], tt.wantBytes) {
				t.Errorf("writeData incorrect data buffer -- got %v expected %v", pb.writtenData[:pb.bytesWritten], tt.wantBytes)
			}
		})
	}
}

func TestPrintfBuffer_Printf(t *testing.T) {
	t.Parallel()
	type pbData struct {
		writtenData  []byte
		bytesWritten int
	}
	type args struct {
		format string
		args   []interface{}
	}
	tests := []struct {
		name    string
		pbData  pbData
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "no expand",
			pbData: pbData{
				writtenData:  make([]byte, 16),
				bytesWritten: 0,
			},
			args: args{
				format: "%s",
				args:   []interface{}{"test"},
			},
			want:    4,
			wantErr: false,
		},
		{
			name: "expand",
			pbData: pbData{
				writtenData:  make([]byte, 16),
				bytesWritten: 16,
			},
			args: args{
				format: "test %s",
				args:   []interface{}{"testtest"},
			},
			want:    13,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := PrintfBuffer{
				writtenData:  tt.pbData.writtenData,
				bytesWritten: tt.pbData.bytesWritten,
			}
			got, err := pb.Printf(tt.args.format, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintfBuffer.Printf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrintfBuffer.Printf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintfBuffer_GetData(t *testing.T) {
	t.Parallel()
	type pbData struct {
		writtenData  []byte
		bytesWritten int
	}
	tests := []struct {
		name   string
		pbData pbData
		want   []byte
	}{
		{
			name: "basic",
			pbData: pbData{
				writtenData:  []byte("abcdefghijklmnop"),
				bytesWritten: 5,
			},
			want: []byte("abcde"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := PrintfBuffer{
				writtenData:  tt.pbData.writtenData,
				bytesWritten: tt.pbData.bytesWritten,
			}
			if got := pb.GetData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrintfBuffer.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintfBuffer_Reset(t *testing.T) {
	t.Parallel()
	type pbData struct {
		writtenData  []byte
		bytesWritten int
	}
	tests := []struct {
		name   string
		pbData pbData
	}{
		{
			name: "already reset",
			pbData: pbData{
				writtenData:  []byte("abcdefghijklmnop"),
				bytesWritten: 0,
			},
		},
		{
			name: "not already reset",
			pbData: pbData{
				writtenData:  []byte("abcdefghijklmnop"),
				bytesWritten: 12,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := PrintfBuffer{
				writtenData:  tt.pbData.writtenData,
				bytesWritten: tt.pbData.bytesWritten,
			}
			pb.Reset()
			if !reflect.DeepEqual(pb.GetData(), []byte{}) {
				t.Errorf("Reset did not result in an empty data array, got %v", pb.GetData())
			}
		})
	}
}
