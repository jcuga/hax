package input

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jcuga/hax/options"
)

func GetInput(opts options.Options) (io.Reader, error) {
	var reader io.Reader
	if len(opts.InputData) > 0 {
		reader = strings.NewReader(opts.InputData)
	} else if len(opts.Filename) > 0 {
		f, err := os.Open(opts.Filename)
		if err != nil {
			return nil, err
		}
		// NOTE: calling code will have to check type of return value
		// and defer close for file.
		reader = f
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	// Now turn reader into reader with given mode...
	var modeReader io.Reader
	switch opts.InputMode {
	case options.Raw:
		// take as-is:
		modeReader = reader
	case options.Hex:
		modeReader = hex.NewDecoder(reader)
	case options.Base64:
		modeReader = base64.NewDecoder(base64.StdEncoding, reader)
	default:
		return nil, fmt.Errorf("Invalid input mode: %v", opts.InputMode)
	}

	return modeReader, nil
}
