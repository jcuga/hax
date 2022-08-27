package output

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func outputHex(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	buf := make([]byte, outBufferSize)
	bytesWritten := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

	var outWriter io.Writer
	outWriter = writer
	if opts.Display.Width > 0 { // wrap to add newlines every width bytes
		// each byte uses 2 chars:
		width := opts.Display.Width * 2
		outWriter, _ = NewFixedWidthWriter(outWriter, width)
	}
	encoder := hex.NewEncoder(outWriter)

	for {
		var n int
		var err error
		// only read up to limit many bytes:
		if opts.Limit-bytesWritten < outBufferSize {
			n, err = reader.Read(buf[:opts.Limit-bytesWritten])
		} else {
			n, err = reader.Read(buf)
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			break
		}

		encoder.Write(buf[0:n])

		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			break
		}
	}
	if !isPipe {
		// add newline to end of terminal output
		fmt.Fprintf(writer, "\n")
	}
}

// TODO: refactor to accept prefix/suffix and use for both str and list funcs
func outputHexString(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	buf := make([]byte, outBufferSize)
	bytesWritten := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

	var outWriter io.Writer
	outWriter = writer
	if opts.Display.Width > 0 { // wrap to add newlines every width bytes
		// each byte takes up 4 chars: len("\xAB") == 4
		width := opts.Display.Width * 4
		outWriter, _ = NewFixedWidthWriter(outWriter, width)
	}

	bytesBuf := bytes.Buffer{}
	for {
		var n int
		var err error
		// only read up to limit many bytes:
		if opts.Limit-bytesWritten < outBufferSize {
			n, err = reader.Read(buf[:opts.Limit-bytesWritten])
		} else {
			n, err = reader.Read(buf)
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			break
		}

		bytesBuf.Reset()
		bytesBuf.Grow(n) // pre-allocate if needed, on subsequent calls, shoudln't allocate

		// TODO: make this faster -- update: faster now... but should be able to get 10-100x better still...
		for i := 0; i < n; i++ {
			highByte := buf[i] >> 4
			if highByte < 10 {
				highByte = '0' + highByte // gets char values '0' thru '9'
			} else {
				highByte = 'A' + (highByte - 10) // gets char values 'A' thru 'F'
			}
			lowByte := buf[i] & 0x0F
			if lowByte < 10 {
				lowByte = '0' + lowByte // gets char values '0' thru '9'
			} else {
				lowByte = 'A' + (lowByte - 10) // gets char values 'A' thru 'F'
			}
			bytesBuf.Write([]byte{'\\', 'x', byte(highByte), lowByte})
		}
		outWriter.Write(bytesBuf.Bytes())

		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			break
		}
	}
	if !isPipe {
		// add newline to end of terminal output
		fmt.Fprintf(writer, "\n")
	}
}
