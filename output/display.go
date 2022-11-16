package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func displayHex(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
	showPretty := !isPipe || opts.Display.Pretty
	subWidthPadding := "  " // if opts.Display.SubWidth set, this amount of whitespace to pad between elements within row
	count := int64(0)
	row := int64(0)
	offsetPadding := int64(0)
	if opts.Offset > 0 {
		offsetPadding = opts.Offset % int64(opts.Display.Width)
	}
	// NOTE: under the hood we're using an input.FixedLengthBufferedReader
	// which will fill up the entire requested buffer on read so we'll
	// get the full opts.Display.Width full of data on all but our last
	// read call which will have partial data before EOF.
	// Without this custom reader type, we don't have a strong guarantee
	// of getting all our requested buffer filled, and in fact, with base64
	// input, the underlying base64 decoder will return often less than
	// the requested size. Thankfully, our cusotm reader will smooth things
	// out so we always get a full line.
	buf := make([]byte, opts.Display.Width)
	wrappedReader := zeroPageOmitter{
		reader:   reader,
		pageSize: opts.Display.PageSize,
		row:      -1, // -1 as on first read ++ will make 0, as we're zero based and do modulo math using this value.
	}
	for {
		// pad/indent offset input so row start counts are nice and aligned.
		// ex: offset 3 --> 3 spaces of padding on first row and row count
		// still says 0 versus no padding and row count starts at 3.
		// So for first row of offset data, grab less than normal.
		var n int
		var err error
		var bytesSkipped int

		if row == 0 && offsetPadding > 0 {
			if opts.Display.OmitZeroPages && opts.Display.PageSize > 0 {
				n, err, bytesSkipped = wrappedReader.ReadRow(buf[:len(buf)-int(offsetPadding)])
			} else {
				n, err = reader.Read(buf[:len(buf)-int(offsetPadding)])
			}
		} else {
			if opts.Display.OmitZeroPages && opts.Display.PageSize > 0 {
				n, err, bytesSkipped = wrappedReader.ReadRow(buf)
			} else {
				n, err = reader.Read(buf)
			}
		}

		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
				os.Exit(1)
			}
		}

		if bytesSkipped > 0 {
			beforeOffset := row*int64(opts.Display.Width) + (opts.Offset - offsetPadding)
			row += int64(bytesSkipped / opts.Display.Width)
			afterOffset := ((row)*int64(opts.Display.Width) + (opts.Offset - offsetPadding)) - 1
			fmt.Fprintf(writer, "%13X-%X omitted. [%d bytes/%d lines of all zeros]\n", beforeOffset, afterOffset, bytesSkipped, bytesSkipped/opts.Display.Width)
			if count+int64(bytesSkipped) >= opts.Limit {
				break
			}
			count += int64(bytesSkipped)
		}

		if n < 1 {
			if row == 0 {
				fmt.Fprintf(writer, "<NO DATA>\n")
				os.Exit(3) // TODO: document return codes
			}
			break
		}

		if row == 0 || (opts.Display.PageSize > 0 && row%int64(opts.Display.PageSize) == int64(0)) {
			fmt.Fprintf(writer, "\n")
			fmt.Fprintf(writer, "%15s", "")
			for i := 0; i < opts.Display.Width; i++ {
				if opts.Display.SubWidth > 0 && i > 0 && i%opts.Display.SubWidth == 0 {
					fmt.Fprintf(writer, "%s", subWidthPadding)
				}

				if showPretty {
					fmt.Fprintf(writer, "%2s ", fmt.Sprintf("\033[36m%2X\033[0m", i))
				} else {
					fmt.Fprintf(writer, "%2X ", i)
				}
			}
			fmt.Fprintf(writer, "\n")

		}

		m := n
		if count+int64(n) > opts.Limit {
			// truncate output of current buffer to not exceed limit
			m = int(opts.Limit - count) // TODO: fix this as only supports int not int64
		}

		count += int64(n)
		rowStart := row*int64(opts.Display.Width) + (opts.Offset - offsetPadding)
		offsetPaddingWhitespace := ""
		if row == 0 && offsetPadding > 0 {
			// NOTE: 3 spaces per offset byte as we have 2 byte hex plus space in between each.
			if opts.Display.SubWidth > 0 {
				offsetPaddingWhitespace = strings.Repeat("   ", int(offsetPadding)) + strings.Repeat(subWidthPadding, (int(offsetPadding)/opts.Display.SubWidth))
			} else {
				offsetPaddingWhitespace = strings.Repeat("   ", int(offsetPadding))
			}
		}
		if showPretty {
			fmt.Fprintf(writer, "\033[36m%13X: \033[0m%s", rowStart, offsetPaddingWhitespace)
		} else {
			fmt.Fprintf(writer, "%13X: %s", rowStart, offsetPaddingWhitespace)
		}

		// Print hex
		for i := 0; i < m; i++ {
			if opts.Display.SubWidth > 0 && i > 0 && ((row != 0 && i%opts.Display.SubWidth == 0) || (row == 0 && (i+int(offsetPadding))%opts.Display.SubWidth == 0)) {
				fmt.Fprintf(writer, "%s", subWidthPadding)
			}
			if buf[i] == 0 && opts.Display.HideZerosBytes {
				fmt.Fprintf(writer, "   ")
			} else {
				fmt.Fprintf(writer, "%02X ", buf[i])
			}
		}
		fmt.Fprintf(writer, "\n%15s%s", "", offsetPaddingWhitespace)
		if !opts.Display.Quiet {
			// Print ascii
			for i := 0; i < m; i++ {
				if opts.Display.SubWidth > 0 && i > 0 && ((row != 0 && i%opts.Display.SubWidth == 0) || (row == 0 && (i+int(offsetPadding))%opts.Display.SubWidth == 0)) {
					fmt.Fprintf(writer, "%s", subWidthPadding)
				}
				// printable ascii only
				var out string
				if buf[i] >= 32 && buf[i] <= 126 {
					out = fmt.Sprintf("%2c", buf[i])
				} else {
					switch buf[i] {
					case 0x09:
						out = fmt.Sprintf("%2s", "\\t")
					case 0x0A:
						out = fmt.Sprintf("%2s", "\\n")
					case 0x0D:
						out = fmt.Sprintf("%2s", "\\r")
					default:
						// just padding
						out = fmt.Sprintf("%2s", "")
					}
				}
				// Don't add bold/colored output if this is piped to
				// another command like less as that will not display nicely.
				if showPretty {
					fmt.Fprintf(writer, "%2s ", fmt.Sprintf("\033[32m%2s\033[0m", out))
				} else {
					fmt.Fprintf(writer, "%2s ", fmt.Sprintf("%2s", out))
				}
			}
		}
		fmt.Fprintf(writer, "\n")

		if count >= opts.Limit {
			break
		}

		row += 1
	}
	fmt.Fprintf(writer, "\n")
}

type zeroPageOmitter struct {
	reader       *input.FixedLengthBufferedReader
	pageSize     int
	row          int64
	rowBuffer    [][]byte
	bytesSkipped int
}

func (z *zeroPageOmitter) ReadRow(buf []byte) (int, error, int) {
	if len(z.rowBuffer) > 0 {
		first := z.rowBuffer[0]
		z.rowBuffer = z.rowBuffer[1:]
		copy(buf, first)
		return len(first), nil, 0
	}
	n, err := z.reader.Read(buf)
	z.row++
	if n > 0 && err == nil && z.row%int64(z.pageSize) == 0 && allZeros(buf[0:n]) {
		// buffer all zeros and keep reading more. clear per page, increment skip count.
		// once (if ever) hit non-zero, start returning any buffered followed by nonzeor
		copied := make([]byte, n)
		copy(copied, buf[0:n])
		z.rowBuffer = append(z.rowBuffer, copied)

		for {
			n, err := z.reader.Read(buf)
			z.row++

			if n > 0 && err == nil && allZeros(buf[0:n]) {

				copied := make([]byte, n)
				copy(copied, buf[0:n])
				z.rowBuffer = append(z.rowBuffer, copied)

				if int(z.row%int64(z.pageSize)) == z.pageSize-1 { // buffered a full page--clear first pageSize many from buffer, increment bytes skipped
					for i := 0; i < z.pageSize; i++ {
						z.bytesSkipped += len(z.rowBuffer[i])
					}
					z.rowBuffer = z.rowBuffer[z.pageSize:]
				}
				continue
			}

			if err != nil && err != io.EOF {
				return 0, err, 0 // report error, forget about any buffered data...
			}

			if n > 0 {
				copied := make([]byte, n)
				copy(copied, buf[0:n])
				z.rowBuffer = append(z.rowBuffer, copied)
			}
			// else: reacehd end--will want to return any buffered data in order.
			// should have at least one item in buffer since added before for loop...
			// in either case (n==0, n>0) we'll return first row of data here.
			first := z.rowBuffer[0]
			z.rowBuffer = z.rowBuffer[1:]
			copy(buf, first)
			skipped := z.bytesSkipped
			z.bytesSkipped = 0
			return len(first), nil, skipped
		}

	} else {
		return n, err, 0
	}
}
func allZeros(buf []byte) bool {
	for _, val := range buf {
		if val != byte(0) {
			return false
		}
	}
	return true
}
