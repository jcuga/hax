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
		seekBufSize := int64(1024 * 10)
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

// Wraps an io.Reader and ignores/omits specific supplied bytes (ex: whitespace).
// Very similar to encoding/base64's internal newlineFilteringReader.
type FilteringReader struct {
	wrapped io.Reader
	// Optional, additional bytes to ignore in addition to whitespace
	ignoreBytes map[byte]struct{}
}

func NewFilteringReader(reader io.Reader, ignore []byte) *FilteringReader {
	stub := struct{}{} // empty struct takes zero space, used as map value when only need a set.
	toIgnore := make(map[byte]struct{})
	// create set of ignored bytes
	for _, b := range ignore {
		toIgnore[b] = stub
	}
	return &FilteringReader{
		wrapped:     reader,
		ignoreBytes: toIgnore,
	}
}

func (r FilteringReader) Read(p []byte) (int, error) {
	n, err := r.wrapped.Read(p)
	for n > 0 {
		offset := 0
		for i, b := range p[:n] {
			if _, found := r.ignoreBytes[b]; !found {
				if i != offset {
					p[offset] = b
				}
				offset++
			}
		}
		if offset > 0 {
			return offset, err
		}
		// Previous buffer entirely ignored bytes, read again
		n, err = r.wrapped.Read(p)
	}
	return n, err
}

// FixedLengthBufferedReader will read/fill the entire requested []byte
// on each read excepting the last one which may yield partial amount.
// For base64 input, calling read(buf) with a buffer of size n will often
// (when len(buf) is not a multiple of 3) yield less than a full buf or data.
// This wrapped reader type will fill up the entire requested buf on a Read
// call unless the EOF has been reached. Client code can then request fixed
// size chunks and receive full data with the exception of the last chunk
// of data in the underlying reader. This helps allow output.displayHex()
// function to request oen full row of data at a time via a simple Read call.
// This may also be useful for other future input formats that may not reliably
// yield a full buffer of data on Read.
type FixedLengthBufferedReader struct {
	wrapped      io.Reader
	buf          []byte
	bufFilledLen int
	bufIndex     int
}

func NewFixedLengthBufferedReader(reader io.Reader) *FixedLengthBufferedReader {
	return &FixedLengthBufferedReader{
		wrapped:      reader,
		buf:          make([]byte, 1024*10),
		bufFilledLen: 0,
		bufIndex:     0,
	}
}

// Read and populate the entirety of input buffer p unless wrapped reader
// has reached EOF in which case p may be partially filled.
// NOTE: has to be pointer-receiver as this function modifies fields!
// Otherwise each call to Read is modifying a copy!
func (r *FixedLengthBufferedReader) Read(p []byte) (int, error) {
	reqLen := len(p)
	bufferedLen := r.bufFilledLen - r.bufIndex
	if bufferedLen >= reqLen {
		copy(p, r.buf[r.bufIndex:r.bufIndex+reqLen])
		r.bufIndex += reqLen
		return reqLen, nil
	}

	if bufferedLen > 0 { // have some amount < reqLen but not zero
		// copy remaining buffered data
		copy(p, r.buf[r.bufIndex:r.bufIndex+bufferedLen])
		r.bufIndex += bufferedLen
	}

	outstandingLen := reqLen - bufferedLen
	alreadyReadLen := bufferedLen

	// used up any buffered data--request more
	for {
		n, err := r.wrapped.Read(r.buf)
		r.bufIndex = 0
		r.bufFilledLen = n

		if n == 0 {
			if err != nil && err != io.EOF {
				// pass along any non-trivial error--for example: if underlying
				// hex/base64 decoder finds invalid data.
				return alreadyReadLen, err
			}
			return alreadyReadLen, io.EOF
		}

		if n >= outstandingLen {
			copy(p[alreadyReadLen:], r.buf[:outstandingLen])
			r.bufIndex += outstandingLen
			return alreadyReadLen + outstandingLen, err
		}

		// read more (n>0), but not enough still
		// take everythign we have and then read more/iterate again
		copy(p[alreadyReadLen:], r.buf[:n])
		r.bufIndex += n
		outstandingLen -= n
		alreadyReadLen += n

		if outstandingLen < 1 {
			return alreadyReadLen, err
		}
	}
}
