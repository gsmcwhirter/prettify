package testutil

import (
	"io"
	"reflect"
	"testing"
)

func TestNewWriterBuffer(t *testing.T) {
	t.Parallel()
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want io.Writer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewWriterBuffer(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWriterBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriterBuffer_expandIfNecessary(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	type args struct {
		addingBytes int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			pb.expandIfNecessary(tt.args.addingBytes)
		})
	}
}

func TestWriterBuffer_writeBytes(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			if err := pb.writeBytes(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("WriterBuffer.writeBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriterBuffer_GetData(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			if got := pb.GetData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriterBuffer.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriterBuffer_Reset(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			pb.Reset()
		})
	}
}

func TestWriterBuffer_Close(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			if err := pb.Close(); (err != nil) != tt.wantErr {
				t.Errorf("WriterBuffer.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriterBuffer_Write(t *testing.T) {
	t.Parallel()
	type fields struct {
		writtenData  []byte
		bytesWritten int
		isClosed     bool
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantBytesWritten int
		wantErr          bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := &WriterBuffer{
				writtenData:  tt.fields.writtenData,
				bytesWritten: tt.fields.bytesWritten,
				isClosed:     tt.fields.isClosed,
			}
			gotBytesWritten, err := pb.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriterBuffer.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBytesWritten != tt.wantBytesWritten {
				t.Errorf("WriterBuffer.Write() = %v, want %v", gotBytesWritten, tt.wantBytesWritten)
			}
		})
	}
}
