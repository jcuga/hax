package commands

import (
	"fmt"
	"io"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

// CountBytes counds the number of bytes from a given input.
// NOTE: the input could be a non-raw format like base64, hex string, etc.
// This will count the "true"/raw amount of bytes represented by the various formats.
func CountBytes(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo, opts options.Options,
	cmdOptions []string) error {
	if len(cmdOptions) != 0 {
		return fmt.Errorf("Command 'count', unexpected arguments. Expect: 0, got: %d.\nUsage: count", len(cmdOptions))
	}

	buf := make([]byte, options.OutputBufferSize)
	bytesRead := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

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
			return fmt.Errorf("Error reading data: %v", err)
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
	return nil
}
