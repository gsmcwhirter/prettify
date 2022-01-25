package testutil

import (
	"errors"
	"io"

	"github.com/gsmcwhirter/prettify/pkg/minmax"
)

// ReadSeeker is a struct implementing io.ReadSeeker for an in-memory byte slice
// This is useful for testing/mocking files
type ReadSeeker struct {
	dataBuffer []byte
	pos        int64
}

// ErrSeekOutOfBounds means that a seek attempted to go beyond the bounds of the data buffer
var ErrSeekOutOfBounds = errors.New("seek out of bounds")

// ErrUnknownSeekWhence means that a Seek had an unknown reference point value
var ErrUnknownSeekWhence = errors.New("unknown Seek whence value")

// ErrReadError is a testing error returned by default from an ErrReadSeeker.Read()
var ErrReadError = errors.New("read error")

// ErrSeekError is a testing error returned fby default from an ErrReadSeeker.Seek()
var ErrSeekError = errors.New("seek error")

// Read attempts to read len(buffer) bytes from the current position
func (rs *ReadSeeker) Read(buffer []byte) (int, error) {
	di := int(rs.pos) - 1
	for bi := 0; bi < len(buffer); bi++ {
		di++
		if di >= len(rs.dataBuffer) {
			return bi, io.EOF
		}

		buffer[bi] = rs.dataBuffer[di]
		rs.pos++
	}

	return len(buffer), nil
}

// Seek moves the read head to the indicated position
func (rs *ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		rs.pos += offset
	case io.SeekStart:
		rs.pos = offset
	case io.SeekEnd:
		rs.pos = int64(len(rs.dataBuffer)) + offset
	default:
		return rs.pos, ErrUnknownSeekWhence
	}

	if rs.pos < 0 {
		rs.pos = 0
		return rs.pos, ErrSeekOutOfBounds
	}

	if rs.pos > int64(len(rs.dataBuffer)) {
		rs.pos = int64(len(rs.dataBuffer))
		return rs.pos, ErrSeekOutOfBounds
	}

	return rs.pos, nil
}

// NewReadSeeker constructs a new ReadSeeker object
func NewReadSeeker(data []byte) *ReadSeeker {
	return &ReadSeeker{
		dataBuffer: data,
		pos:        0,
	}
}

// ErrReadSeekerConfig is a configuration struct for NewErrReadSeeker
type ErrReadSeekerConfig struct {
	ReadError     error
	SeekError     error
	ErrOnReadCall int
	ErrOnSeekCall int
	FillByte      byte
}

// ErrReadSeeker is a mock for a read call that throws an error on the nth read (and returns ReadBytes bytes otherwise)
type ErrReadSeeker struct {
	ReadError     error
	SeekError     error
	ErrOnReadCall int
	ErrOnSeekCall int
	ReadBytes     int
	TotalBytes    int
	FillByte      byte
	readCalls     int
	seekCalls     int
	pos           int64
}

// NewErrReadSeeker creates a new ErrReadSeeker from the config and some defaults
func NewErrReadSeeker(readBytes, totalBytes int, ersConfig ErrReadSeekerConfig) *ErrReadSeeker {
	if ersConfig.ReadError == nil {
		ersConfig.ReadError = ErrReadError
	}

	if ersConfig.SeekError == nil {
		ersConfig.SeekError = ErrSeekError
	}

	if ersConfig.FillByte == 0 {
		ersConfig.FillByte = byte(' ')
	}

	return &ErrReadSeeker{
		ReadError:     ersConfig.ReadError,
		SeekError:     ersConfig.SeekError,
		ErrOnReadCall: ersConfig.ErrOnReadCall,
		ErrOnSeekCall: ersConfig.ErrOnSeekCall,
		ReadBytes:     readBytes,
		TotalBytes:    totalBytes,
		FillByte:      ersConfig.FillByte,
		readCalls:     0,
		seekCalls:     0,
		pos:           0,
	}
}

// Read returns some bytes in buffer or raises an error
func (ers *ErrReadSeeker) Read(buffer []byte) (int, error) {
	if ers.readCalls == ers.ErrOnReadCall {
		ers.readCalls++
		return 0, ers.ReadError
	}

	ers.readCalls++

	rbytes := minmax.IntMin(ers.ReadBytes, len(buffer))
	rbytes = minmax.IntMin(rbytes, ers.TotalBytes-int(ers.pos))
	for i := 0; i < rbytes; i++ {
		buffer[i] = ers.FillByte
	}

	ers.pos += int64(rbytes)

	return rbytes, nil
}

// Seek moves the read pointer or raises an error
func (ers *ErrReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if ers.seekCalls == ers.ErrOnSeekCall {
		ers.seekCalls++
		return ers.pos, ers.SeekError
	}

	ers.seekCalls++

	switch whence {
	case io.SeekCurrent:
		ers.pos += offset
	case io.SeekStart:
		ers.pos = offset
	case io.SeekEnd:
		ers.pos = int64(ers.TotalBytes) + offset
	default:
		return ers.pos, ErrUnknownSeekWhence
	}

	if ers.pos < 0 {
		ers.pos = 0
		return ers.pos, ErrSeekOutOfBounds
	}

	if ers.pos > int64(ers.TotalBytes) {
		ers.pos = int64(ers.TotalBytes)
		return ers.pos, ErrSeekOutOfBounds
	}

	return ers.pos, nil
}
