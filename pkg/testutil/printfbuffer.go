package testutil

import "fmt"

// PrintfBuffer can stand in for fmt.Printf when testing
type PrintfBuffer struct {
	writtenData  []byte
	bytesWritten int
}

// NewPrintfBuffer creates a new, empty print buffer with the given starting size
func NewPrintfBuffer(size int) PrintfBuffer {
	return PrintfBuffer{
		writtenData:  make([]byte, size),
		bytesWritten: 0,
	}
}

func (pb *PrintfBuffer) expandIfNecessary(addingBytes int) {
	for pb.bytesWritten+addingBytes > len(pb.writtenData) {
		newBuffer := make([]byte, 2*len(pb.writtenData))
		for i := 0; i < len(pb.writtenData); i++ {
			newBuffer[i] = pb.writtenData[i]
		}
		pb.writtenData = newBuffer
	}
}

func (pb *PrintfBuffer) writeBytes(data []byte) {
	pb.expandIfNecessary(len(data))
	for i, byt := range data {
		pb.writtenData[pb.bytesWritten+i] = byt
	}
	pb.bytesWritten += len(data)
}

// Printf writes the requested data into the buffer
func (pb *PrintfBuffer) Printf(format string, args ...interface{}) (int, error) {
	newBytes := []byte(fmt.Sprintf(format, args...))
	pb.writeBytes(newBytes)
	return len(newBytes), nil
}

// GetData returns the data written so far
func (pb *PrintfBuffer) GetData() []byte {
	return pb.writtenData[:pb.bytesWritten]
}

// Reset lazily resets the buffer (does not erase existing data, just resets a pointer)
func (pb *PrintfBuffer) Reset() {
	pb.bytesWritten = 0
}
