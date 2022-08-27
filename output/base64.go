package output

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func outputBase64(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	buf := make([]byte, outBufferSize)
	bytesWritten := int64(0) // num input bytes written, NOT the number of bytes the base64 output fills.
	var outWriter io.Writer
	outWriter = writer
	if opts.Display.Width > 0 { // wrap to add newlines every width bytes
		outWriter, _ = NewFixedWidthWriter(outWriter, opts.Display.Width)
	}
	encoder := base64.NewEncoder(base64.StdEncoding, outWriter)
	defer func() {
		encoder.Close() // needed to flush/encode any final, partial block of data
		if !isPipe {
			// add newline to end of terminal output
			fmt.Fprintf(writer, "\n")
		}
	}()

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
