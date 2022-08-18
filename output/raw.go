package output

import (
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

const (
	rawOutBufferSize = 1024 * 100 // TODO: 100kb buffer--adequate?
)

func outputRaw(reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	// TODO: if isPipe, prompt first? cmdline -y option?
	buf := make([]byte, rawOutBufferSize)

	bytesWritten := int64(0)
	for {
		var n int
		var err error
		// only read up to limit many bytes:
		if opts.Limit-bytesWritten < rawOutBufferSize {
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

		os.Stdout.Write(buf[0:n])
		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			return
		}
	}
}
