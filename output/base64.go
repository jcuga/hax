package output

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func outputBase64(reader *input.FixedLengthBufferedReader, opts options.Options) {
	buf := make([]byte, outBufferSize)
	bytesWritten := int64(0) // num input bytes written, NOT the number of bytes the base64 output fills.
	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	defer encoder.Close() // needed to flush/encode any final, partial block of data

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
			return
		}
	}
}
