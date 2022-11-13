package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

// CountBytes counds the number of bytes from a given input.
// NOTE: the input could be a non-raw format like base64, hex string, etc.
// This will count the "true"/raw amount of bytes represented by the various formats.
func CountBytes(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options,
	cmdOptions []string) {
	if len(cmdOptions) != 0 {
		fmt.Fprintf(os.Stderr, "Command 'count', unexpected arguments. Expect: 0, got: %d.\n", len(cmdOptions))
		fmt.Fprint(os.Stderr, "Usage: count\n")
		os.Exit(1)
	}

	buf := make([]byte, options.OutputBufferSize)
	bytesRead := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

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
		if opts.Limit-bytesRead < options.OutputBufferSize {
			n, err = reader.Read(buf[:opts.Limit-bytesRead])
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

		bytesRead += int64(n)
		if bytesRead >= opts.Limit {
			break
		}
	}

	if !opts.Display.Quiet {
		fmt.Fprintf(writer, "%d bytes", bytesRead)
		kb := float64(bytesRead) / 1024
		mb := float64(bytesRead) / (1024 * 1024)
		gb := float64(bytesRead) / (1024 * 1024 * 1024)
		if gb >= 1.0 {
			fmt.Fprintf(writer, "\n%0.2f GB", gb)
		} else if mb >= 1.0 {
			fmt.Fprintf(writer, "\n%0.2f MB", mb)
		} else if kb > 0.1 {
			fmt.Fprintf(writer, "\n%0.2f KB", kb)
		}
	} else {
		fmt.Fprintf(writer, "%d", bytesRead)
	}
}
