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
		// The std hex decoder expects only 2-digit hex chars, no whitespace.
		// Wrap in a reader that filters out whitespace.
		modeReader = hex.NewDecoder(&whitespaceFilteringReader{reader})
	case options.Base64:
		modeReader = base64.NewDecoder(base64.StdEncoding, &whitespaceFilteringReader{reader})
	default:
		return nil, fmt.Errorf("Invalid input mode: %v", opts.InputMode)
	}

	return modeReader, nil
}

// Wraps an io.Reader and ignores/omits whitespace.
// Very similar to encoding/base64's internal newlineFilteringReader except
// this also ignores tab and space.
type whitespaceFilteringReader struct {
	wrapped io.Reader
}

func (r whitespaceFilteringReader) Read(p []byte) (int, error) {
	n, err := r.wrapped.Read(p)
	for n > 0 {
		offset := 0
		for i, b := range p[:n] {
			if b != '\r' && b != '\n' && b != '\t' && b != ' ' {
				if i != offset {
					p[offset] = b
				}
				offset++
			}
		}
		if offset > 0 {
			return offset, err
		}
		// Previous buffer entirely whitespace, read again
		n, err = r.wrapped.Read(p)
	}
	return n, err
}
