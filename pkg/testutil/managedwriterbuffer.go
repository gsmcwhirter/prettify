package testutil

import "io"

// WriterBuffer can stand in for fmt.Printf when testing
type WriterBuffer struct {
	writtenData  []byte
	bytesWritten int
	isClosed     bool
}

// NewWriterBuffer creates a new, empty print buffer with the given starting size
func NewWriterBuffer(size int) io.Writer {
	return &WriterBuffer{
		writtenData:  make([]byte, size),
		bytesWritten: 0,
		isClosed:     false,
	}
}

func (pb *WriterBuffer) expandIfNecessary(addingBytes int) {
	for pb.bytesWritten+addingBytes > len(pb.writtenData) {
		newBuffer := make([]byte, 2*len(pb.writtenData))
		for i := 0; i < len(pb.writtenData); i++ {
			newBuffer[i] = pb.writtenData[i]
		}
		pb.writtenData = newBuffer
	}
}

func (pb *WriterBuffer) writeBytes(data []byte) error {
	if pb.isClosed {
		return io.ErrClosedPipe
	}
	pb.expandIfNecessary(len(data))
	for i, byt := range data {
		pb.writtenData[pb.bytesWritten+i] = byt
	}
	pb.bytesWritten += len(data)
	return nil
}

// GetData returns the data written so far
func (pb *WriterBuffer) GetData() []byte {
	return pb.writtenData[:pb.bytesWritten]
}

// Reset lazily resets the buffer (does not erase existing data, just resets a pointer)
func (pb *WriterBuffer) Reset() {
	pb.bytesWritten = 0
	pb.isClosed = false
}

// Close sets the buffer to "closed" state
func (pb *WriterBuffer) Close() error {
	pb.isClosed = true
	return nil
}

func (pb *WriterBuffer) Write(data []byte) (int, error) {
	if err := pb.writeBytes(data); err != nil {
		return 0, err
	}
	return len(data), nil
}
