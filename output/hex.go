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
		outWidth := opts.Display.Width * 2
		outWriter, _ = NewFixedWidthWriter(outWriter, outWidth)
	}
	encoder := hex.NewEncoder(outWriter)

	defer func() {
		if !isPipe {
			// add newline to end of terminal output
			fmt.Fprintf(writer, "\n")
		}
	}()

	if !isPipe {
		// add newline to start of output when in terminal
		fmt.Fprintf(writer, "\n")
	}

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
}

func outputHexStringOrList(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	buf := make([]byte, outBufferSize)
	bytesWritten := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

	var outWriter io.Writer
	outWriter = writer
	var outWidth int
	if opts.Display.Width > 0 { // wrap to add newlines every width bytes
		if opts.OutputMode == options.HexString {
			// each byte takes up 4 chars: len("\xAB") == 4
			outWidth = opts.Display.Width * 4
		} else { // options.HexList
			// each byte takes up 6 chars: len("0xAB, ") == 6
			outWidth = opts.Display.Width * 6
		}
		outWriter, _ = NewFixedWidthWriter(outWriter, outWidth)
	}

	bytesBuf := bytes.Buffer{}
	bytesBuf.Grow(outWidth) // pre-allocate if needed, on subsequent calls, shouldn't allocate
	firstByte := true       // decides if ", " included in front of hex list items

	defer func() {
		if !isPipe {
			// add newline to end of terminal output
			fmt.Fprintf(writer, "\n")
		}
	}()

	if !isPipe {
		// add newline to start of output when in terminal
		fmt.Fprintf(writer, "\n")
	}

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

			if opts.OutputMode == options.HexString {
				bytesBuf.Write([]byte{'\\', 'x', byte(highByte), lowByte})
			} else { // options.HexList
				if firstByte {
					bytesBuf.Write([]byte{'0', 'x', byte(highByte), lowByte})
					firstByte = false
				} else {
					bytesBuf.Write([]byte{',', ' ', '0', 'x', byte(highByte), lowByte})
				}
			}
		}
		outWriter.Write(bytesBuf.Bytes())

		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			break
		}
	}
}
