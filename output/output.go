package output

import (
	"bufio"
	"fmt"
	"io"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

const (
	outBufferSize = 1024 * 100 // TODO: 100kb buffer--adequate?
)

func Output(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) error {
	// use buffered writer for better performance.
	// ex: displaying line by line to stdout or a file when displaying hex is slow.
	// buffering the writes significantly speeds up the hex display output.
	w := bufio.NewWriter(writer)
	defer w.Flush()
	switch opts.OutputMode {
	case options.Base64:
		outputBase64(w, reader, isPipe, opts)
	case options.Display:
		displayHex(w, reader, isPipe, opts)
	case options.Hex:
		outputHex(w, reader, isPipe, opts)
	case options.HexString:
		outputHexStringOrList(w, reader, isPipe, opts)
	case options.HexList:
		outputHexStringOrList(w, reader, isPipe, opts)
	case options.Raw:
		outputRaw(w, reader, isPipe, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}
