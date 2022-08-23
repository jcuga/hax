package input

import (
	"io"
)

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
		buf:          make([]byte, readerBufferSize),
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
