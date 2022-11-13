package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

// TODO: optional offset
// TODO: optional min widht?
// TODO: optional coloring of offset?

func outputStrings(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	showPretty := !isPipe || opts.Display.Pretty
	buf := make([]byte, outBufferSize)
	bytesRead := int64(0) // num input bytes written, NOT the number of bytes the hex output fills.

	// buffer output
	var outBuilder strings.Builder
	// buffer current string--may omit if all whitespace or too short
	var curStrBuilder strings.Builder
	// track start of current string
	curStringStart := int64(-1)

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
		if opts.Limit-bytesRead < outBufferSize {
			n, err = reader.Read(buf[:opts.Limit-bytesRead])
		} else {
			n, err = reader.Read(buf)
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty)
			fmt.Fprint(writer, outBuilder.String())
			outBuilder.Reset()
			break
		}

		for i := 0; i < n; i++ {
			if buf[i] > 31 && buf[i] < 127 {
				if curStringStart < 0 { // not set
					curStringStart = opts.Offset + bytesRead + int64(i)
				}
				curStrBuilder.WriteByte(buf[i])
			} else {
				flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty)
			}
		}

		fmt.Fprint(writer, outBuilder.String())
		outBuilder.Reset()

		bytesRead += int64(n)
		if bytesRead >= opts.Limit {
			flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty)
			fmt.Fprint(writer, outBuilder.String())
			outBuilder.Reset()
			break
		}
	}
}

func flushCurString(curStrBuilder, outBuilder *strings.Builder, opts *options.Options, curStringStart *int64, showPretty bool) {
	if curStrBuilder.Len() > 0 {
		orig := curStrBuilder.String()
		trimmed := strings.TrimSpace(orig)
		if len(trimmed) >= opts.Display.MinStringLen && len(trimmed) <= opts.Display.MaxStringLen {
			// account for any preceeding whitespace when showing offset to start of displayed string
			if len(orig) > len(trimmed) {
				idx := strings.IndexByte(orig, trimmed[0])
				if idx > 0 {
					*curStringStart += int64(idx)
				}
			}
			if !opts.Display.Quiet {
				if showPretty {
					outBuilder.WriteString(fmt.Sprintf("\033[36m%13X:\t\033[0m", *curStringStart))
				} else {
					outBuilder.WriteString(fmt.Sprintf("%13X:\t", *curStringStart))
				}
			}
			outBuilder.WriteString(trimmed)
			outBuilder.WriteByte('\n')
		}
		*curStringStart = -1
		curStrBuilder.Reset()
	}
}
