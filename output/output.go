package output

import (
	"fmt"
	"io"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

const (
	outBufferSize = 1024 * 100 // TODO: 100kb buffer--adequate?
)

func Output(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) error {
	switch opts.OutputMode {
	case options.Base64:
		outputBase64(writer, reader, isPipe, opts)
	case options.Display:
		displayHex(writer, reader, isPipe, opts)
	case options.Hex:
		outputHex(writer, reader, isPipe, opts)
	case options.HexString:
		outputHexStringOrList(writer, reader, isPipe, opts)
	case options.HexList:
		outputHexStringOrList(writer, reader, isPipe, opts)
	case options.Raw:
		outputRaw(writer, reader, isPipe, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}
