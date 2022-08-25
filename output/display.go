package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func displayHex(reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) {
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
	for {
		// pad/indent offset input so row start counts are nice and aligned.
		// ex: offset 3 --> 3 spaces of padding on first row and row count
		// still says 0 versus no padding and row count starts at 3.
		// So for first row of offset data, grab less than normal.
		var n int
		var err error
		if row == 0 && offsetPadding > 0 {
			n, err = reader.Read(buf[:len(buf)-int(offsetPadding)])
		} else {
			n, err = reader.Read(buf)
		}

		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
				os.Exit(1)
			}
		}

		if n < 1 {
			if row == 0 {
				fmt.Printf("<NO DATA>\n")
				os.Exit(3) // TODO: document return codes
			}

			break
		}

		if row == 0 {
			fmt.Println("")
			fmt.Printf("%15s", "")
			for i := 0; i < opts.Display.Width; i++ {
				if opts.Display.SubWidth > 0 && i > 0 && i%opts.Display.SubWidth == 0 {
					fmt.Printf("%s", subWidthPadding)
				}

				if showPretty {
					fmt.Printf("%2s ", fmt.Sprintf("\033[36m%2X\033[0m", i))
				} else {
					fmt.Printf("%2X ", i)
				}
			}
			fmt.Println("")

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
			offsetPaddingWhitespace = strings.Repeat("   ", int(offsetPadding))
		}
		if showPretty {
			fmt.Printf("\033[36m%13X: \033[0m%s", rowStart, offsetPaddingWhitespace)
		} else {
			fmt.Printf("%13X: %s", rowStart, offsetPaddingWhitespace)
		}

		// Print hex
		for i := 0; i < m; i++ {
			if opts.Display.SubWidth > 0 && i > 0 && i%opts.Display.SubWidth == 0 {
				fmt.Printf("%s", subWidthPadding)
			}
			fmt.Printf("%02X ", buf[i])
		}
		fmt.Printf("\n%15s%s", "", offsetPaddingWhitespace)
		if !opts.Display.NoAscii {
			// Print ascii
			for i := 0; i < m; i++ {
				if opts.Display.SubWidth > 0 && i > 0 && i%opts.Display.SubWidth == 0 {
					fmt.Printf("%s", subWidthPadding)
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
					fmt.Printf("%2s ", fmt.Sprintf("\033[32m%2s\033[0m", out))
				} else {
					fmt.Printf("%2s ", fmt.Sprintf("%2s", out))
				}
			}
		}
		fmt.Println("")
		if count >= opts.Limit {
			break
		}

		// NOTE: checking && m == opts.Display.Width to avoid printing after a partial line as that implies the subsequent
		// read will be EOF.
		if opts.Display.PageSize > 0 && row%int64(opts.Display.PageSize) == (int64(opts.Display.PageSize)-1) && m == opts.Display.Width {
			fmt.Println("")
			fmt.Printf("%15s", "")
			for i := 0; i < opts.Display.Width; i++ {
				if opts.Display.SubWidth > 0 && i > 0 && i%opts.Display.SubWidth == 0 {
					fmt.Printf("%s", subWidthPadding)
				}

				if showPretty {
					fmt.Printf("%2s ", fmt.Sprintf("\033[36m%2X\033[0m", i))
				} else {
					fmt.Printf("%2X ", i)
				}
			}
			fmt.Println("")
		}
		row += 1
	}
	fmt.Println("")
}
