package output

import (
	"fmt"
	"io"
)

// Wraps an io.Writer and inserts newlines after every N bytes of output.
type FixedWidthWriter struct {
	wrapped   io.Writer
	lineWidth int
	counter   int
}

func NewFixedWidthWriter(writer io.Writer, width int) (*FixedWidthWriter, error) {
	if width < 1 {
		return nil, fmt.Errorf("width must be >= 1, got: %d", width)
	}
	return &FixedWidthWriter{
		wrapped:   writer,
		lineWidth: width,
		counter:   0,
	}, nil
}

func (w *FixedWidthWriter) Write(p []byte) (n int, err error) {
	remaining := len(p)
	for {
		if remaining == 0 {
			break
		}
		if w.counter == w.lineWidth {
			w.wrapped.Write([]byte{'\n'})
			w.counter = 0
		}
		// write min(remaining, width-counter), as width-counter is how much left on cur line
		min := w.lineWidth - w.counter
		if remaining < min {
			min = remaining
		}
		offset := len(p) - remaining
		n, err := w.wrapped.Write(p[offset : offset+min])
		if err != nil {
			return n, err
		}
		if n != min {
			return n, fmt.Errorf("Unexpected write result: %d, expected: %d", n, min)
		}
		remaining -= min
		w.counter += min
	}
	return len(p), nil
}
