package output

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/jcuga/hax/options"
)

func Output(reader io.Reader, isPipe bool, opts options.Options) error {
	// TODO: fix base64 and hex intput that has newlines
	// TODO: add support for base64 and hex output with optional width
	// TODO: prevent raw out to char device (or prompt y/n?)
	switch opts.OutputMode {
	case options.Display:
		displayHex(reader, isPipe, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}

// TODO: clean up this func
// TODO: idea: only have hex not dec+hex, but put dec on ascii line
func displayHex(reader io.Reader, isPipe bool, opts options.Options) {
	showPretty := !isPipe || opts.Display.Pretty
	if opts.Limit <= 0 {
		opts.Limit = math.MaxInt64
	}
	fmt.Println("")
	fmt.Printf("%19s", "")
	for i := 0; i < opts.Display.Width; i++ {
		if showPretty {
			fmt.Printf("%2s ", fmt.Sprintf("\033[36m%2X\033[0m", i))
		} else {
			fmt.Printf("%2X ", i)
		}
	}
	fmt.Println("")

	count := int64(0)
	row := int64(0)
	buf := make([]byte, opts.Display.Width)
	fmt.Println("")
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading data: %v\n", err)
				os.Exit(1)
			}
		}
		if n < 1 {
			break
		}
		m := n
		if count+int64(n) > opts.Limit {
			// truncate output of current buffer to not exceed limit
			m = int(opts.Limit - count) // TODO: fix this as only supports int not int64
		}
		count += int64(n)

		rowStart := row*int64(opts.Display.Width) + opts.Offset
		if showPretty {
			fmt.Printf("\033[36m%8d %8X: \033[0m", rowStart, rowStart)
		} else {
			fmt.Printf("%8d %8X: ", rowStart, rowStart)
		}

		// Print hex
		for i := 0; i < m; i++ {
			if showPretty {
				// fmt.Printf("%02X ", buf[i])
				fmt.Printf("\033[1m%02X\033[0m ", buf[i])
			} else {
				fmt.Printf("%02X ", buf[i])
			}
		}
		fmt.Printf("\n%19s", "")
		if !opts.Display.NoAscii {
			// Print ascii
			for i := 0; i < m; i++ {
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
		row += 1
		if count >= opts.Limit {
			break
		}

		if opts.Display.PageSize > 0 && row%int64(opts.Display.PageSize) == (int64(opts.Display.PageSize)-1) {
			fmt.Println("")
			fmt.Printf("%19s", "")
			for i := 0; i < opts.Display.Width; i++ {
				if showPretty {
					fmt.Printf("%2s ", fmt.Sprintf("\033[36m%2X\033[0m", i))
				} else {
					fmt.Printf("%2X ", i)
				}
			}
			fmt.Println("")
		}

	}
	fmt.Println("")
}
