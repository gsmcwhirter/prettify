package testutil

import (
	"io"
	"reflect"
	"testing"
)

func testReadSeeker(data []byte, pos int64) *ReadSeeker {
	rs := NewReadSeeker(data)
	rs.pos = pos
	return rs
}

func testErrReadSeeker(readBytes, totalBytes int, ersConfig ErrReadSeekerConfig, readCalls, seekCalls int, pos int64) *ErrReadSeeker {
	ers := NewErrReadSeeker(readBytes, totalBytes, ersConfig)
	ers.readCalls = readCalls
	ers.seekCalls = seekCalls
	ers.pos = pos
	return ers
}

func TestReadSeeker_Read(t *testing.T) {
	t.Parallel()
	type args struct {
		buffer []byte
	}
	tests := []struct {
		name         string
		rs           *ReadSeeker
		args         args
		wantNumBytes int
		wantErr      bool
	}{
		{
			name: "small target buffer",
			rs:   NewReadSeeker([]byte("abcdefgh")),
			args: args{
				buffer: make([]byte, 4),
			},
			wantNumBytes: 4,
			wantErr:      false,
		},
		{
			name: "medium target buffer",
			rs:   NewReadSeeker([]byte("abcdefgh")),
			args: args{
				buffer: make([]byte, 8),
			},
			wantNumBytes: 8,
			wantErr:      false,
		},
		{
			name: "large target buffer",
			rs:   NewReadSeeker([]byte("abcdefgh")),
			args: args{
				buffer: make([]byte, 16),
			},
			wantNumBytes: 8,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotNumBytes, err := tt.rs.Read(tt.args.buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadSeeker.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNumBytes != tt.wantNumBytes {
				t.Errorf("ReadSeeker.Read() = %v, want %v", gotNumBytes, tt.wantNumBytes)
			}
		})
	}
}

func TestReadSeeker_Seek(t *testing.T) {
	t.Parallel()
	type args struct {
		offset int64
		whence int
	}
	tests := []struct {
		name       string
		rs         *ReadSeeker
		args       args
		wantNewPos int64
		wantErr    bool
	}{
		{
			name: "seek beginning",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 0,
				whence: io.SeekStart,
			},
			wantNewPos: 0,
			wantErr:    false,
		},
		{
			name: "seek beginning pos",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 3,
				whence: io.SeekStart,
			},
			wantNewPos: 3,
			wantErr:    false,
		},
		{
			name: "seek beginning neg",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: -1,
				whence: io.SeekStart,
			},
			wantNewPos: 0,
			wantErr:    true,
		},
		{
			name: "seek beginning pos too far",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 7,
				whence: io.SeekStart,
			},
			wantNewPos: 4,
			wantErr:    true,
		},
		{
			name: "seek end",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 0,
				whence: io.SeekEnd,
			},
			wantNewPos: 4,
			wantErr:    false,
		},
		{
			name: "seek end pos",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 1,
				whence: io.SeekEnd,
			},
			wantNewPos: 4,
			wantErr:    true,
		},
		{
			name: "seek end neg",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: -1,
				whence: io.SeekEnd,
			},
			wantNewPos: 3,
			wantErr:    false,
		},
		{
			name: "seek end neg too far",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: -5,
				whence: io.SeekEnd,
			},
			wantNewPos: 0,
			wantErr:    true,
		},
		{
			name: "seek current",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 0,
				whence: io.SeekCurrent,
			},
			wantNewPos: 2,
			wantErr:    false,
		},
		{
			name: "seek current pos",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 2,
				whence: io.SeekCurrent,
			},
			wantNewPos: 4,
			wantErr:    false,
		},
		{
			name: "seek current neg",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: -1,
				whence: io.SeekCurrent,
			},
			wantNewPos: 1,
			wantErr:    false,
		},
		{
			name: "seek current pos too far",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 4,
				whence: io.SeekCurrent,
			},
			wantNewPos: 4,
			wantErr:    true,
		},
		{
			name: "seek current neg too far",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: -3,
				whence: io.SeekCurrent,
			},
			wantNewPos: 0,
			wantErr:    true,
		},
		{
			name: "bad whence",
			rs:   testReadSeeker([]byte("test"), 2),
			args: args{
				offset: 0,
				whence: -1,
			},
			wantNewPos: 2,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotNewPos, err := tt.rs.Seek(tt.args.offset, tt.args.whence)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadSeeker.Seek() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNewPos != tt.wantNewPos {
				t.Errorf("ReadSeeker.Seek() = %v, want %v", gotNewPos, tt.wantNewPos)
			}
		})
	}
}

func TestNewReadSeeker(t *testing.T) {
	t.Parallel()
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want *ReadSeeker
	}{
		{
			name: "basic test",
			args: args{
				data: []byte("abcd"),
			},
			want: &ReadSeeker{
				dataBuffer: []byte("abcd"),
				pos:        0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewReadSeeker(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReadSeeker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrReadSeeker(t *testing.T) {
	t.Parallel()
	type args struct {
		readBytes  int
		totalBytes int
		ersConfig  ErrReadSeekerConfig
	}
	tests := []struct {
		name string
		args args
		want *ErrReadSeeker
	}{
		{
			name: "basic test",
			args: args{
				readBytes:  10,
				totalBytes: 20,
				ersConfig:  ErrReadSeekerConfig{},
			},
			want: &ErrReadSeeker{
				ReadError:     ErrReadError,
				SeekError:     ErrSeekError,
				ErrOnReadCall: 0,
				ErrOnSeekCall: 0,
				ReadBytes:     10,
				TotalBytes:    20,
				FillByte:      byte(' '),
				readCalls:     0,
				seekCalls:     0,
				pos:           0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewErrReadSeeker(tt.args.readBytes, tt.args.totalBytes, tt.args.ersConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrReadSeeker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrReadSeeker_Read(t *testing.T) {
	t.Parallel()
	type result struct {
		bytesRead int
		wantErr   bool
	}
	type args struct {
		buffer []byte
	}
	tests := []struct {
		name        string
		ers         *ErrReadSeeker
		args        args
		wantResults []result
	}{
		{
			name: "err on first read",
			ers:  testErrReadSeeker(10, 20, ErrReadSeekerConfig{}, 0, 0, 0),
			args: args{
				buffer: make([]byte, 12),
			},
			wantResults: []result{
				{0, true},
				{10, false},
				{10, false},
				{0, false},
			},
		},
		{
			name: "err on second read",
			ers:  testErrReadSeeker(10, 20, ErrReadSeekerConfig{ErrOnReadCall: 1}, 0, 0, 0),
			args: args{
				buffer: make([]byte, 12),
			},
			wantResults: []result{
				{10, false},
				{0, true},
				{10, false},
				{0, false},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i, res := range tt.wantResults {
				i, res := i, res
				got, err := tt.ers.Read(tt.args.buffer)
				if (err != nil) != res.wantErr {
					t.Errorf("ErrReadSeeker.Read() %d error = %v, wantErr %v", i, err, res.wantErr)
					return
				}
				if got != res.bytesRead {
					t.Errorf("ErrReadSeeker.Read() %d = %v, want %v", i, got, res.bytesRead)
				}
			}
		})
	}
}

func TestErrReadSeeker_Seek(t *testing.T) {
	t.Parallel()
	type result struct {
		wantNewPos int64
		wantErr    bool
	}
	type args struct {
		offset int64
		whence int
	}
	tests := []struct {
		name        string
		ers         *ErrReadSeeker
		args        args
		wantResults []result
	}{
		{
			name: "seek beginning",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 0,
				whence: io.SeekStart,
			},
			wantResults: []result{
				{0, false},
				{0, true},
				{0, false},
			},
		},
		{
			name: "seek beginning pos",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 3,
				whence: io.SeekStart,
			},
			wantResults: []result{
				{3, false},
				{3, true},
				{3, false},
			},
		},
		{
			name: "seek beginning neg",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: -1,
				whence: io.SeekStart,
			},
			wantResults: []result{
				{0, true},
				{0, true},
				{0, true},
			},
		},
		{
			name: "seek beginning pos too far",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 25,
				whence: io.SeekStart,
			},
			wantResults: []result{
				{20, true},
				{20, true},
				{20, true},
			},
		},
		{
			name: "seek end",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 0,
				whence: io.SeekEnd,
			},
			wantResults: []result{
				{20, false},
				{20, true},
				{20, false},
			},
		},
		{
			name: "seek end pos",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 1,
				whence: io.SeekEnd,
			},
			wantResults: []result{
				{20, true},
				{20, true},
				{20, true},
			},
		},
		{
			name: "seek end neg",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: -1,
				whence: io.SeekEnd,
			},
			wantResults: []result{
				{19, false},
				{19, true},
				{19, false},
			},
		},
		{
			name: "seek end neg too far",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: -25,
				whence: io.SeekEnd,
			},
			wantResults: []result{
				{0, true},
				{0, true},
				{0, true},
			},
		},
		{
			name: "seek current",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 0,
				whence: io.SeekCurrent,
			},
			wantResults: []result{
				{2, false},
				{2, true},
				{2, false},
			},
		},
		{
			name: "seek current pos",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 2,
				whence: io.SeekCurrent,
			},
			wantResults: []result{
				{4, false},
				{4, true},
				{6, false},
			},
		},
		{
			name: "seek current neg",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 10),
			args: args{
				offset: -1,
				whence: io.SeekCurrent,
			},
			wantResults: []result{
				{9, false},
				{9, true},
				{8, false},
			},
		},
		{
			name: "seek current pos too far",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 14),
			args: args{
				offset: 4,
				whence: io.SeekCurrent,
			},
			wantResults: []result{
				{18, false},
				{18, true},
				{20, true},
			},
		},
		{
			name: "seek current neg too far",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 6),
			args: args{
				offset: -4,
				whence: io.SeekCurrent,
			},
			wantResults: []result{
				{2, false},
				{2, true},
				{0, true},
			},
		},
		{
			name: "bad whence",
			ers:  testErrReadSeeker(1, 20, ErrReadSeekerConfig{ErrOnSeekCall: 1}, 0, 0, 2),
			args: args{
				offset: 0,
				whence: -1,
			},
			wantResults: []result{
				{2, true},
				{2, true},
				{2, true},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i, res := range tt.wantResults {
				i, res := i, res
				gotNewPos, err := tt.ers.Seek(tt.args.offset, tt.args.whence)
				if (err != nil) != res.wantErr {
					t.Errorf("ErrReadSeeker.Seek() %d error = %v, wantErr %v", i, err, res.wantErr)
					return
				}
				if gotNewPos != res.wantNewPos {
					t.Errorf("ErrReadSeeker.Seek() %d = %v, want %v", i, gotNewPos, res.wantNewPos)
				}
			}
		})
	}
}
