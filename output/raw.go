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

		// For TTY/terminal (ie not a pipe) output, warn and ask if first batch looks like non-printables
		if !isPipe && bytesWritten == 0 && containsNonPrintable(buf[:n]) {
			if !opts.Yes {
				if !promptForYes("Output to character device contains non-printable bytes.") { // TODO: better wording--see curl for example? IIRC does similar.
					os.Exit(0) // TODO: or a non-zero, distinct exit code? see how other common tools do it too.
				}
			}
		}

		os.Stdout.Write(buf[0:n])
		bytesWritten += int64(n)
		if bytesWritten >= opts.Limit {
			return
		}
	}
}

// TODO: move to common/util module if needed elsewhere
func containsNonPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return true
		}
	}
	return false
}

// TODO: move to common/util module if needed elsewhere
func promptForYes(msg string) bool {
	fmt.Fprintf(os.Stderr, "%s Are you sure? (Y/y): ", msg)
	var answer string
	fmt.Scanln(&answer)
	if len(answer) > 0 && (answer[0] == 'y' || answer[0] == 'Y') {
		return true
	}
	return false
}