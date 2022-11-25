package commands

import (
	"fmt"
	"io"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/jcuga/hax/eval"
	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func StringsUtf8(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo, opts options.Options,
	cmdOptions []string) error {
	minStringLen := 0
	maxStringLen := math.MaxInt32
	if len(cmdOptions) > 2 {
		return fmt.Errorf("Too many arguments for command 'utf8', expect 0-2, got: %d\nUsage: utf8 [minLen] [maxLen]", len(cmdOptions))
	}
	if len(cmdOptions) > 1 {
		if parsedMax, err := eval.ParseHexDecOrBin(cmdOptions[1]); err == nil {
			maxStringLen = int(parsedMax)
		} else {
			return fmt.Errorf("Command: 'utf8', failed to parse max length arg: %q, err: %v\nUsage: utf8 [minLen] [maxLen]",
				cmdOptions[1], err)
		}
	}
	if len(cmdOptions) > 0 {
		if parsedMin, err := eval.ParseHexDecOrBin(cmdOptions[0]); err == nil {
			minStringLen = int(parsedMin)
		} else {
			return fmt.Errorf("Command: 'utf8', failed to parse min length arg: %q, err: %v\nUsage: utf8 [minLen] [maxLen]",
				cmdOptions[0], err)
		}
	}
	if minStringLen > maxStringLen || maxStringLen < 1 {
		return fmt.Errorf("Command: 'utf8', invalid min/max len args. min: %d, max: %d. Must have min < max and max > 0\nUsage: utf8 [minLen] [maxLen]",
			minStringLen, maxStringLen)
	}

	buf := make([]byte, options.OutputBufferSize)
	bytesRead := int64(0)

	// buffer output
	var outBuilder strings.Builder
	state := stringsUtf8State{
		runeBuffer: make([]byte, 0, 4),
	}
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
			break
		}

		for i := 0; i < n; i++ {
			if state.bytesRemaining > 0 { // in middle of unicode sequence
				if buf[i]>>6 == 0b10 { // starts with continuation--appears valid
					state.bytesRemaining -= 1
					state.runeBuffer = append(state.runeBuffer, buf[i])
					if state.bytesRemaining == 0 {
						// add complete unicode rune to str builder, but DONT call flush
						r, _ := utf8.DecodeRune(state.runeBuffer)
						if r != utf8.RuneError {
							state.curStrBuilder.WriteRune(r)
						} else {
							// wasn't valid, omit and flush any existing str data
							state.flush(&outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
						}
						state.runeBuffer = state.runeBuffer[:0]
					}
					continue
				} else {
					// must not be unicode--flush anything we have and reset.
					state.flush(&outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
				}
				// NOTE: can still use current byte (buf[i]) later below if it looks like
				// printable ascii or the start of a unicode rune
			}

			foundSequenceStart := false
			if buf[i]>>5 == 0b110 { // start of 2 byte sequence
				state.runeBuffer = append(state.runeBuffer, buf[i])
				state.bytesRemaining = 1
				foundSequenceStart = true
			} else if buf[i]>>4 == 0b1110 { // start of 3 byte sequence
				state.runeBuffer = append(state.runeBuffer, buf[i])
				state.bytesRemaining = 2
				foundSequenceStart = true
			} else if buf[i]>>3 == 0b11110 { // start of 4 byte sequence
				state.runeBuffer = append(state.runeBuffer, buf[i])
				state.bytesRemaining = 3
				foundSequenceStart = true
			}
			if foundSequenceStart {
				if curStringStart < 0 { // not set
					curStringStart = opts.Offset + bytesRead + int64(i)
				}
				continue
			}

			// wasn't the middle or start of a unicode sequence
			// but could still be ascii/trivial unicode
			// NOTE: only consuming printable ascii
			// considering non-printables to be not part of a string.
			// so any non-printable would split our output.
			if buf[i] > 31 && buf[i] < 127 {
				if curStringStart < 0 { // not set
					curStringStart = opts.Offset + bytesRead + int64(i)
				}
				state.curStrBuilder.WriteByte(buf[i])
			} else {
				state.flush(&outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
			}
		}

		if outBuilder.Len() > 0 {
			fmt.Fprint(writer, outBuilder.String())
			outBuilder.Reset()
		}

		bytesRead += int64(n)
		if bytesRead >= opts.Limit {
			break
		}

	}

	state.flush(&outBuilder, &first, &opts, &curStringStart, ioInfo.OutputPretty, minStringLen, maxStringLen)
	if outBuilder.Len() > 0 {
		fmt.Fprint(writer, outBuilder.String())
		outBuilder.Reset()
	}
	return nil
}

type stringsUtf8State struct {
	curStrBuilder strings.Builder
	// How many bytes left to consume when buffering current rune
	bytesRemaining int
	runeBuffer     []byte
}

func (s *stringsUtf8State) flush(outBuilder *strings.Builder, first *bool, opts *options.Options, curStringStart *int64, showPretty bool, minStringLen, maxStringLen int) {
	// NOTE: not bothering with contents of s.runeBuffer as flush() is called
	// when either finished consuming bytes/no more data, OR determined that
	// next byte was invalid. So any partial data is not usable at this point.
	if s.curStrBuilder.Len() > 0 {
		orig := s.curStrBuilder.String()
		trimmed := strings.TrimSpace(orig)
		asRunes := []rune(trimmed)
		if len(asRunes) > 0 && len(asRunes) >= minStringLen && len(asRunes) <= maxStringLen {
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
		s.curStrBuilder.Reset()
	}
	*curStringStart = -1
	s.runeBuffer = s.runeBuffer[:0]
	s.bytesRemaining = 0
}
