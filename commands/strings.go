package commands

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/jcuga/hax/eval"
	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Strings(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo, opts options.Options,
	cmdOptions []string) error {
	minStringLen := 0
	maxStringLen := math.MaxInt32
	if len(cmdOptions) > 2 {
		return fmt.Errorf("Too many arguments for command 'strings', expect 0-2, got: %d\nUsage: strings [minLen] [maxLen]", len(cmdOptions))
	}
	if len(cmdOptions) > 1 {
		if parsedMax, err := eval.ParseHexDecOrBin(cmdOptions[1]); err == nil {
			maxStringLen = int(parsedMax)
		} else {
			return fmt.Errorf("Command: 'strings', failed to parse max length arg: %q, err: %v\nUsage: strings [minLen] [maxLen]",
				cmdOptions[1], err)
		}
	}
	if len(cmdOptions) > 0 {
		if parsedMin, err := eval.ParseHexDecOrBin(cmdOptions[0]); err == nil {
			minStringLen = int(parsedMin)
		} else {
			return fmt.Errorf("Command: 'strings', failed to parse min length arg: %q, err: %v\nUsage: strings [minLen] [maxLen]",
				cmdOptions[0], err)
		}
	}
	if minStringLen > maxStringLen || maxStringLen < 1 {
		return fmt.Errorf("Command: 'strings', invalid min/max len args. min: %d, max: %d. Must have min < max and max > 0\nUsage: strings [minLen] [maxLen]",
			minStringLen, maxStringLen)
	}

	buf := make([]byte, options.OutputBufferSize)
	bytesRead := int64(0)

	// buffer output
	var outBuilder strings.Builder
	// buffer current string--may omit if all whitespace or too short
	var curStrBuilder strings.Builder
	// track start of current string
	curStringStart := int64(-1)
	first := true // used to know when to omit preceeding newline as first line doesn't need it

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
			flushCurString(&curStrBuilder, &outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
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
				flushCurString(&curStrBuilder, &outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
			}
		}

		fmt.Fprint(writer, outBuilder.String())
		outBuilder.Reset()

		bytesRead += int64(n)
		if bytesRead >= opts.Limit {
			flushCurString(&curStrBuilder, &outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
			fmt.Fprint(writer, outBuilder.String())
			outBuilder.Reset()
			break
		}
	}
	return nil
}

func flushCurString(curStrBuilder, outBuilder *strings.Builder, first *bool, opts *options.Options, curStringStart *int64, showPretty bool,
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
			if !*first {
				outBuilder.WriteByte('\n')
			}
			if !opts.Display.Quiet {
				if showPretty {
					outBuilder.WriteString(fmt.Sprintf("\033[36m%13X:\t\033[0m", *curStringStart))
				} else {
					outBuilder.WriteString(fmt.Sprintf("%13X:\t", *curStringStart))
				}
			}
			outBuilder.WriteString(trimmed)
			*first = false
		}
		curStrBuilder.Reset()
	}
	*curStringStart = -1
}
