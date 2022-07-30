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

// For base64 input, calling read(buf) with a buffer of size n will often
// (when len(buf) is not a multiple of 3) yield less than a full buf or data.
// This wrapped reader type will fill up the entire requested buf on a Read
// call unless the EOF has been reached. Client code can then request fixed
// size chunks and receive full data with the exception of the last chunk
// of data in the underlying reader. This helps allow output.displayHex()
// function to request oen full row of data at a time via a simple Read call.
// This may also be useful for other future input formats that may not reliably
// yield a full buffer of data on Read.
type fixedLengthBufferedReader struct {
	wrapped      io.Reader
	buf          []byte
	bufFilledLen int
	bufIndex     int
}

func NewFixedLengthBufferedReader(reader io.Reader) *fixedLengthBufferedReader {
	return &fixedLengthBufferedReader{
		wrapped:      reader,
		buf:          make([]byte, 1024*5),
		bufFilledLen: 0,
		bufIndex:     0,
	}
}

// Read and populate the entirety of input buffer p unless wrapped reader
// has reached EOF in which case p may be partially filled.
// NOTE: has to be pointer-receiver as this function modifies fields!
// Otherwise each call to Read is modifying a copy!
func (r *fixedLengthBufferedReader) Read(p []byte) (int, error) {
	fmt.Printf("READ called with req buf len: %d\n", len(p))
	reqLen := len(p)
	bufferedLen := r.bufFilledLen - r.bufIndex
	fmt.Printf("bufferedLen: %d, r.bufFilledLen: %d, r.bufIndex: %d\n", bufferedLen, r.bufFilledLen, r.bufIndex)
	if bufferedLen >= reqLen {
		copy(p, r.buf[r.bufIndex:r.bufIndex+reqLen])
		r.bufIndex += reqLen
		fmt.Printf("RETURN-A reqLen: %d\n", reqLen)
		return reqLen, nil
	}

	if bufferedLen > 0 { // have some amount < reqLen but not zero
		// copy remaining buffered data
		fmt.Printf("Use existing bufferedLen: %d\n", bufferedLen)
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
		fmt.Printf("read more internally, n: %d\n", n)

		if n == 0 {
			fmt.Printf("RETURN-B, ret: %d, err: %v\n", alreadyReadLen, err)
			return alreadyReadLen, io.EOF
		}

		if n >= outstandingLen {
			fmt.Printf("copy using n: %d, alreadyReadLen: %d, outstandingLen: %d, return: %d, err: %v\n", n, alreadyReadLen, outstandingLen, (alreadyReadLen + outstandingLen), err)
			copy(p[alreadyReadLen:], r.buf[:outstandingLen])
			r.bufIndex += outstandingLen
			fmt.Printf("r.bufIndex: %d, r.bufFilledLen: %d\n", r.bufIndex, r.bufFilledLen)
			fmt.Printf("RETURN-C, ret: %d, err: %v\n", alreadyReadLen+outstandingLen, err)
			return alreadyReadLen + outstandingLen, err
		}

		// read more (n>0), but not enough still
		// take everythign we have and then read more/iterate again
		copy(p[alreadyReadLen:], r.buf[:n])
		r.bufIndex += n
		outstandingLen -= n
		alreadyReadLen += n

		if outstandingLen < 1 {
			fmt.Printf("RETURN-D outstandingLen: %d, returnign alreadyLen: %d, err: %v\n", outstandingLen, alreadyReadLen, err)
			return alreadyReadLen, err
		}
	}
}
