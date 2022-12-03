package commands

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

const (
	anyByte = 256
)

func Search(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo, opts options.Options,
	cmdOptions []string) error {
	if len(cmdOptions) < 1 || len(cmdOptions) > 2 {
		return errors.New("Usage: search <pattern> [b:a]\noptional b:a is num bytes before/after to display on matches, ex: search \\x01\\x02\\x03 5:7")
	}
	beforeBytes, afterBytes := 8, 8
	var err error
	if len(cmdOptions) == 2 {
		beforeBytes, afterBytes, err = parseBeforeAfter(cmdOptions[1])
		if err != nil {
			return err
		}
	}

	fmt.Printf("pattern raw: %q\n", cmdOptions[0])

	s, err := NewSearcher(cmdOptions[0], beforeBytes, afterBytes)
	if err != nil {
		return err
	}
	if s != nil { // TODO: remove me

	}
	return nil // TODO: build me
}

// parseBeforeAfter takes a string of form "<int>:<int>" where the ints
// are >= 0 lengths before after a match to buffer/display.
func parseBeforeAfter(input string) (int, int, error) {
	idx := strings.IndexByte(input, ':')
	if idx == -1 || idx == len(input)-1 || idx == 0 {
		return 0, 0, fmt.Errorf("expected format: int:int, got: %q", input)
	}
	before, err := strconv.Atoi(input[:idx])
	if err != nil {
		return 0, 0, err
	}
	after, err := strconv.Atoi(input[idx+1:])
	if err != nil {
		return 0, 0, err
	}
	if before < 0 || after < 0 {
		return 0, 0, fmt.Errorf("values cannot be negative, got: %d and %d", before, after)
	}
	return before, after, nil
}

// NewSearcher creates a searcher with a given byte pattern parsed from inputPattern.
func NewSearcher(inputPattern string, showBeforeBytes int, showAfterBytes int) (*searcher, error) {
	if len(inputPattern) == 0 {
		return nil, errors.New("empty pattern")
	}
	s := searcher{
		pattern:         make([]uint16, 0),
		showBeforeBytes: showBeforeBytes,
		showAfterBytes:  showAfterBytes,
	}

	for i := 0; i < len(inputPattern); i++ {
		// handle escaped sequences
		if inputPattern[i] == '\\' {
			if i < len(inputPattern)-1 {
				switch inputPattern[i+1] {
				case 'n', 'N':
					s.pattern = append(s.pattern, uint16('\n'))
				case 't', 'T':
					s.pattern = append(s.pattern, uint16('\t'))
				case 'r', 'R':
					s.pattern = append(s.pattern, uint16('\r'))
				case '?': // escaped '?' as normally that's a wildcard/match-any-byte
					s.pattern = append(s.pattern, uint16('?'))
				case 'x', 'X':
					if i < len(inputPattern)-3 {
						hexStr := string([]byte{inputPattern[i+2], inputPattern[i+3]})
						if parsedByte, err := strconv.ParseUint(hexStr, 16, 8); err == nil {
							s.pattern = append(s.pattern, uint16(parsedByte))
							// consume additional 2 bytes (consuming 2nd byte 'x' happens further below...)
							i += 2
						} else {
							return nil, fmt.Errorf("'Invalid hex sequence: '\\x%s'", hexStr)
						}
					} else {
						return nil, errors.New("'\\x' without trailing 2 char hex")
					}
				case '\\':
					s.pattern = append(s.pattern, uint16('\\'))
				default:
					return nil, fmt.Errorf("Invalid escape sequence: '%s'", string([]byte{inputPattern[i], inputPattern[i+1]}))
				}
				i += 1 // consume 2nd byte
			} else {
				return nil, errors.New("Trailing '\\'")
			}
		} else if inputPattern[i] == '?' { // handle wildcard/match-any-single-byte
			s.pattern = append(s.pattern, anyByte)
		} else { // take char as-is
			s.pattern = append(s.pattern, uint16(inputPattern[i]))
		}
	}
	// pre-allocate capacity to match lenth of search pattern as we'll buffer
	// the same amount while building a match
	s.matchBuffer = make([]byte, 0, len(s.pattern))
	return &s, nil
}

// searchMatch represents any matched search pattern.
// Includes any contextual before/after bytes as desired.
type searchMatch struct {
	startIndex   int
	endIndex     int
	matchedValue []byte
	// Optional before contex to display
	beforeBytes []byte
	// Optional after contex to display
	afterBytes []byte
}

// searcher is used to find a given pattern of bytes (with optional match-any-byte placeholder(s))
// and captures all non-overlapping hits (first match wins).
type searcher struct {
	pattern []uint16 // bigger than byte to allow placeholder for "any"
	// buffer current match state in order to match across chunks
	matchBuffer []byte
	matches     []searchMatch
	// How much context before to buffer for a match for informational purposes
	showBeforeBytes int
	// How mcuh context after to buffer for a match for informational purposes.
	showAfterBytes int
	// How many bytes already consumed via calls to searcher.update(inChunk).
	// Used to calculate absolute match start/end indexes--not just relative to
	// currently processed chunk.
	bytesConsumed int
}

// update consumes given chunk of data and checks for matches.
// This can be called multiple times in a row as we'll want to
// pass streamed/chunked data to it.  It will track matches based on the index
// of ALL data passed in, not just the current inChunk.  For example, if
// 100 bytes were passed to update, and then a match occurs on the 2nd byte of
// the subsequent call, the match start would be at index 101 (102nd, as 0 based...)
// not the index 1 (2nd byte) of the current chunk.
func (s *searcher) update(inChunk []byte) {
	patternPos := len(s.matchBuffer)
	for i := 0; i < len(inChunk); i++ {
		if s.pattern[patternPos] == anyByte || uint16(inChunk[i]) == s.pattern[patternPos] {
			s.matchBuffer = append(s.matchBuffer, inChunk[i])
			patternPos++
			if len(s.matchBuffer) == len(s.pattern) {
				copied := make([]byte, len(s.matchBuffer))
				copy(copied, s.matchBuffer)
				newMatch := searchMatch{
					matchedValue: copied,
					endIndex:     s.bytesConsumed + i,
					startIndex:   (s.bytesConsumed + i) - (len(copied) - 1),
				}
				s.matches = append(s.matches, newMatch)
				s.matchBuffer = s.matchBuffer[:0]
				patternPos = 0
			}
		} else {
			// TODO: instead of clearning completely, see if partial overlap match
			s.matchBuffer = s.matchBuffer[:0]
			patternPos = 0
		}
	}
	s.bytesConsumed += len(inChunk)
}

// TODO: figure out how to get pre-post context data efficiently
// TODO: ability to take/consume hits that are found and contextualized so more streaming like?
// NOTE: could be found, but not consumed trailing bytes yet for context...
// TODO: then a final flush() if found near end with not enough for desired trailing context?
