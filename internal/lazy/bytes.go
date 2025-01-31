package lazy

import (
	"bufio"
	"bytes"
	"fmt"
)

// A [Writer] that will lazy initialize itself upon first Write
// and is backed by a [bytes.Buffer].
type Writer struct {
	initialized bool
	buffer      *bytes.Buffer
	writer      *bufio.Writer
	// total bytes written
	totalBytes int
	// the last error
	lastError error
}

func (sw *Writer) Bytes() (b []byte, e error) {
	if !sw.initialized {
		return b, fmt.Errorf("not initialized")
	}

	if sw.lastError != nil {
		return b, fmt.Errorf("last write erorr: %w", sw.lastError)
	}

	flushErr := sw.writer.Flush()
	if flushErr != nil {
		return b, fmt.Errorf("failed to flush internal writer: %w", flushErr)
	}

	return sw.buffer.Bytes(), nil
}

func (sw *Writer) Write(p []byte) (n int, err error) {
	if !sw.initialized {
		sw.buffer = new(bytes.Buffer)
		sw.writer = bufio.NewWriter(sw.buffer)
		sw.initialized = true
	}

	n, err = sw.writer.Write(p)
	sw.totalBytes += n
	sw.lastError = err
	return
}
