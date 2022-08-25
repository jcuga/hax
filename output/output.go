package output

import (
	"fmt"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

const (
	outBufferSize = 1024 * 100 // TODO: 100kb buffer--adequate?
)

func Output(reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) error {
	switch opts.OutputMode {
	case options.Base64:
		outputBase64(reader, isPipe, opts)
	case options.Display:
		displayHex(reader, isPipe, opts)
	case options.Hex:
		outputHex(reader, isPipe, opts)
	case options.Raw:
		outputRaw(reader, isPipe, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}
