package commands

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/jcuga/hax/eval"
	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Strings(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options,
	cmdOptions []string) {
	minStringLen := 0
	maxStringLen := math.MaxInt32
	if len(cmdOptions) > 2 {
		fmt.Fprintf(os.Stderr, "Too many arguments for command 'strings', expect 0-2, got: %d\n", len(cmdOptions))
		fmt.Fprint(os.Stderr, "Usage: strings [minLen] [maxLen]\n")
		os.Exit(1)
	}
	if len(cmdOptions) > 1 {
		if parsedMax, err := eval.ParseHexDecOrBin(cmdOptions[1]); err == nil {
			maxStringLen = int(parsedMax)
		} else {
			fmt.Fprintf(os.Stderr, "Command: 'strings', failed to parse max length arg: %q, err: %v\n",
				cmdOptions[1], err)
			fmt.Fprint(os.Stderr, "Usage: strings [minLen] [maxLen]\n")
			os.Exit(1)
		}
	}
	if len(cmdOptions) > 0 {
		if parsedMin, err := eval.ParseHexDecOrBin(cmdOptions[0]); err == nil {
			minStringLen = int(parsedMin)
		} else {
			fmt.Fprintf(os.Stderr, "Command: 'strings', failed to parse min length arg: %q, err: %v\n",
				cmdOptions[0], err)
			fmt.Fprint(os.Stderr, "Usage: strings [minLen] [maxLen]\n")
			os.Exit(1)
		}
	}
	if minStringLen > maxStringLen || maxStringLen < 1 {
		fmt.Fprintf(os.Stderr, "Command: 'strings', invalid min/max len args. min: %d, max: %d. Must have min < max and max > 0\n",
			minStringLen, maxStringLen)
		fmt.Fprint(os.Stderr, "Usage: strings [minLen] [maxLen]\n")
		os.Exit(1)
	}

	showPretty := !isPipe || opts.Display.Pretty
	buf := make([]byte, options.OutputBufferSize)
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
			flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty, minStringLen, maxStringLen)
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
				flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty, minStringLen, maxStringLen)
			}
		}

		fmt.Fprint(writer, outBuilder.String())
		outBuilder.Reset()

		bytesRead += int64(n)
		if bytesRead >= opts.Limit {
			flushCurString(&curStrBuilder, &outBuilder, &opts, &curStringStart, showPretty, minStringLen, maxStringLen)
			fmt.Fprint(writer, outBuilder.String())
			outBuilder.Reset()
			break
		}
	}
}

func flushCurString(curStrBuilder, outBuilder *strings.Builder, opts *options.Options, curStringStart *int64, showPretty bool,
	minStringLen, maxStringLen int) {
	if curStrBuilder.Len() > 0 {
		orig := curStrBuilder.String()
		trimmed := strings.TrimSpace(orig)
		if len(trimmed) > 0 && len(trimmed) >= minStringLen && len(trimmed) <= maxStringLen {
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
