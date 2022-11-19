package output

import (
	"fmt"
	"io"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func outputRaw(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo, opts options.Options) {
	buf := make([]byte, options.OutputBufferSize)
	bytesWritten := int64(0)

	for {
		var n int
		var err error
		// only read up to limit many bytes:
		if opts.Limit-bytesWritten < options.OutputBufferSize {
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

		// For TTY/terminal (ie not a pipe) output, warn and ask if first batch looks like non-printables
		if !ioInfo.StdoutIsPipe && bytesWritten == 0 && containsNonPrintable(buf[:n]) {
			if !opts.Yes {
				if !promptForYes("Output may be a binary file.  See it anyway?") { // TODO: better wording--see curl for example? IIRC does similar.
					os.Exit(0) // TODO: or a non-zero, distinct exit code? see how other common tools do it too.
				}
			}
		}

		writer.Write(buf[0:n])
		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			return
		}
	}
}

// TODO: move to common/util module if needed elsewhere
// NOTE: allowing printable chars plus 0x07-0x0D (\a thru \r)
// TODO: technically, non printable is <32 and >126, but newline/tab/CR also fine as non "raw" binary files will have these.
func containsNonPrintable(data []byte) bool {
	for _, b := range data {
		if b > 126 || (b < 32 && !(b >= 0x07 && b <= 0x0D)) {
			return true
		}
	}
	return false
}

// TODO: move to common/util module if needed elsewhere
func promptForYes(msg string) bool {
	fmt.Fprintf(os.Stderr, "%s (y/n): ", msg)
	var answer string
	fmt.Scanln(&answer)
	if len(answer) > 0 && (answer[0] == 'y' || answer[0] == 'Y') {
		return true
	}
	return false
}
