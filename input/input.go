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

const (
	readerBufferSize = 1024 * 10
)

// Gets variety of input (stdin, string, file) with format (base64, hex, raw)
// and returns as a FixedLengthBufferedReader which will get n bytes
// for each read([n]byte) call aside from the least/EOF one which may return
// less.  This allows us to easily read and get back expected amount of
// data without having to worry about base64 vs hex vs raw which all have
// different input vs represented byte lengths.
func GetInput(opts options.Options) (*FixedLengthBufferedReader, io.Closer, error) {
	var reader io.Reader
	var closer io.Closer
	// flag for whether we Seek on a raw formatted file to enforce opts.Offset
	// since this is going to be much faster than our read loop further below.
	// For raw file input (not file but base64 or hex contents),
	// seek is the way to go.  When we're not a flle, or a file with a non-raw
	// format, we rely on wrapped readers to ensure the proper amount of
	// "true" bytes are skipped since we're dealing with base64/hex
	// "synthetic" bytes input.
	fileOffsetOptimization := false

	if len(opts.InputData) > 0 {
		reader = strings.NewReader(opts.InputData)
		closer = nil
	} else if len(opts.Filename) > 0 {
		f, err := os.Open(opts.Filename)
		if err != nil {
			return nil, nil, err
		}
		reader = f
		closer = f
		if opts.Offset > 0 && opts.InputMode == options.Raw {
			if _, err := f.Seek(opts.Offset, os.SEEK_SET); err != nil {
				defer f.Close()
				return nil, nil, fmt.Errorf("Failed to seek offset: %d on input file, error: %v", opts.Offset, err)
			}
			fileOffsetOptimization = true
		}
	} else {
		reader = bufio.NewReader(os.Stdin)
		closer = nil
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
		// Now also ignoring "\x" to support byte literal strings like: "\xaa\xbb\xcc\xdd"
		modeReader = hex.NewDecoder(NewFilteringReader(reader, []byte{
			'\r', '\n', '\t', ' ', '\\', 'x',
		}))
	case options.Base64:
		modeReader = base64.NewDecoder(base64.StdEncoding, NewFilteringReader(reader, []byte{
			'\r', '\n', '\t', ' ',
		}))
	default:
		return nil, closer, fmt.Errorf("Invalid input mode: %v", opts.InputMode)
	}

	fixedReader := NewFixedLengthBufferedReader(modeReader)

	if opts.Offset > 0 && !fileOffsetOptimization {
		// Lazy seek--just read enough data and discard.
		seekBufSize := int64(readerBufferSize)
		curOffset := int64(0)
		fill := make([]byte, seekBufSize)
		for curOffset != opts.Offset {
			remaining := opts.Offset - curOffset
			if remaining < seekBufSize {
				fixedReader.Read(fill[:remaining])
				curOffset += remaining
			} else {
				fixedReader.Read(fill)
				curOffset += seekBufSize
			}
		}
	}

	return fixedReader, closer, nil
}
