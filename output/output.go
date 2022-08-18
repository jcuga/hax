package output

import (
	"fmt"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Output(reader *input.FixedLengthBufferedReader, isPipe bool, opts options.Options) error {
	// TODO: add support for base64 and hex output with optional width
	// TODO: prevent raw out to char device (or prompt y/n?)
	switch opts.OutputMode {
	case options.Display:
		displayHex(reader, isPipe, opts)
	case options.Raw:
		outputRaw(reader, isPipe, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}
