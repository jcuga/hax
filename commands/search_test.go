package commands

import (
	"reflect"
	"testing"
)

func Test_searcher_update_matchAcrossCalls(t *testing.T) {
	s, err := NewSearcher("ab", 2, 3)
	if err != nil {
		t.Fatalf("Unexpected err creating searcher: %s", err)
	}

	// Search finds value ("ab") split across two update calls:
	s.update([]byte("hello"))
	s.update([]byte("xyza"))
	s.update([]byte("bqrst"))
	if len(s.matches) != 1 {
		t.Fatalf("Expected %d matches, got: %d", 1, len(s.matches))
	}
	if !reflect.DeepEqual(s.matches[0].matchedValue, []byte("ab")) {
		t.Errorf("Unexpected match value, expect: %q, got: %q", "ab", s.matches[0].matchedValue)
	}
	expectedStart, expectedEnd, expectedConsumed := 8, 9, 14
	if s.matches[0].startIndex != expectedStart {
		t.Errorf("Unexpected startIndex value, expect: %d, got: %d", expectedStart, s.matches[0].startIndex)
	}
	if s.matches[0].endIndex != expectedEnd {
		t.Errorf("Unexpected endIndex value, expect: %d, got: %d", expectedEnd, s.matches[0].endIndex)
	}
	if s.bytesConsumed != expectedConsumed {
		t.Errorf("Unexpected bytesConsumed value, expect: %d, got: %d", expectedConsumed, s.bytesConsumed)
	}
}

func Test_searcher_update_anyByteMatches(t *testing.T) {
	// NOTE: the 2nd ? is an escaped/literal question mark, not an anyByte
	s, err := NewSearcher("\xAAb?\\?", 2, 3)
	if err != nil {
		t.Fatalf("Unexpected err creating searcher: %s", err)
	}

	// Finds 2 matches, note the last bit doesn't match because
	// we have a literal ? (via \\?) not an anyByte
	s.update([]byte("xx\xAAbc?xx\xAAb\x99?xxx\xAAbzyaa"))
	if len(s.matches) != 2 {
		t.Fatalf("Expected %d matches, got: %d", 2, len(s.matches))
	}

	if !reflect.DeepEqual(s.matches[0].matchedValue, []byte("\xAAbc?")) {
		t.Errorf("Unexpected match value, expect: %q, got: %q", "ab", s.matches[0].matchedValue)
	}
	expectedStart, expectedEnd, expectedConsumed := 2, 5, 21
	if s.matches[0].startIndex != expectedStart {
		t.Errorf("Unexpected startIndex value, expect: %d, got: %d", expectedStart, s.matches[0].startIndex)
	}
	if s.matches[0].endIndex != expectedEnd {
		t.Errorf("Unexpected endIndex value, expect: %d, got: %d", expectedEnd, s.matches[0].endIndex)
	}

	if !reflect.DeepEqual(s.matches[1].matchedValue, []byte("\xAAb\x99?")) {
		t.Errorf("Unexpected match value, expect: %q, got: %q", "ab", s.matches[1].matchedValue)
	}
	expectedStart, expectedEnd = 8, 11
	if s.matches[1].startIndex != expectedStart {
		t.Errorf("Unexpected startIndex value, expect: %d, got: %d", expectedStart, s.matches[1].startIndex)
	}
	if s.matches[1].endIndex != expectedEnd {
		t.Errorf("Unexpected endIndex value, expect: %d, got: %d", expectedEnd, s.matches[1].endIndex)
	}

	if s.bytesConsumed != expectedConsumed {
		t.Errorf("Unexpected bytesConsumed value, expect: %d, got: %d", expectedConsumed, s.bytesConsumed)
	}
}

// TODO: overlap
// TODO: overlap more often because of wild card?

func Test_parseBeforeAfter(t *testing.T) {
	type testCase struct {
		input          string
		expectedBefore int
		expectedAfter  int
		expectedErrStr string
	}
	cases := []testCase{
		{"0:0", 0, 0, ""},
		{"0:1", 0, 1, ""},
		{"1:0", 1, 0, ""},
		{"1:1", 1, 1, ""},
		{"123:456", 123, 456, ""},
		{":1", 0, 0, "expected format: int:int, got: \":1\""},
		{"1:", 0, 0, "expected format: int:int, got: \"1:\""},
		{":", 0, 0, "expected format: int:int, got: \":\""},
		{"asdf", 0, 0, "expected format: int:int, got: \"asdf\""},
		{"", 0, 0, "expected format: int:int, got: \"\""},
		{"2:-3", 0, 0, "values cannot be negative, got: 2 and -3"},
		{"-4:5", 0, 0, "values cannot be negative, got: -4 and 5"},
		{"a:2", 0, 0, "strconv.Atoi: parsing \"a\": invalid syntax"},
		{"3:b", 0, 0, "strconv.Atoi: parsing \"b\": invalid syntax"},
		{"1a:2b", 0, 0, "strconv.Atoi: parsing \"1a\": invalid syntax"},
	}
	for _, c := range cases {
		b, a, e := parseBeforeAfter(c.input)
		if e == nil && c.expectedErrStr != "" {
			t.Errorf("input: %q, err nil, but expected: %q", c.input, c.expectedErrStr)
		}
		if e != nil && e.Error() != c.expectedErrStr {
			t.Errorf("input: %q, unexpected err, expected: %q, got: %q", c.input, c.expectedErrStr, e.Error())
		}
		if b != c.expectedBefore {
			t.Errorf("input: %q, unexpected before value, expected: %d, got: %d", c.input, c.expectedBefore, b)
		}
		if a != c.expectedAfter {
			t.Errorf("input: %q, unexpected after value, expected: %d, got: %d", c.input, c.expectedAfter, a)
		}
	}
}

func Test_NewSearcher(t *testing.T) {
	type testCase struct {
		inputPattern     string
		inputBeforeBytes int
		inputAfterBytes  int
		expectedPattern  []uint16
		expectedErrStr   string
	}
	cases := []testCase{
		{"hello", 3, 5, []uint16{'h', 'e', 'l', 'l', 'o'}, ""},
		// raw bytes input mixed with ascii:
		{"\x01\x03bye\x05", 3, 5, []uint16{'\x01', '\x03', 'b', 'y', 'e', '\x05'}, ""},
		// escaped bytes (how user on cmdline most likely to enter:)
		{"\\x01\\x03bye\\x05", 3, 5, []uint16{'\x01', '\x03', 'b', 'y', 'e', '\x05'}, ""},
		{"", 3, 5, nil, "empty pattern"},
		{"a\\b", 3, 5, nil, "Invalid escape sequence: '\\b'"},
		// to do a literal slash, double escape it:
		{"a\\\\b", 3, 5, []uint16{'a', '\\', 'b'}, ""},
		{"a\\\\b\\", 3, 5, nil, "Trailing '\\'"},
		{"a\\\\b\\\\", 3, 5, []uint16{'a', '\\', 'b', '\\'}, ""},

		// ? is a placeholder for anyByte
		{"a?b", 3, 5, []uint16{'a', anyByte, 'b'}, ""},
		// escape ? to use it literally
		{"a\\?b", 3, 5, []uint16{'a', '?', 'b'}, ""},
		{"1\\N\\n2\\R\\r3\\t\\T4\\xff5\\Xaa", 3, 5,
			[]uint16{'1', '\n', '\n', '2', '\r', '\r', '3', '\t', '\t', '4', '\xff', '5', '\xaa'}, ""},
		{"\x0a\x22\x333", 3, 5,
			[]uint16{'\n', '\x22', '\x33', '3'}, ""},
		{"a\\x", 3, 5, nil, "'\\x' without trailing 2 char hex"},
		{"a\\x1", 3, 5, nil, "'\\x' without trailing 2 char hex"},
		{"a\\x12", 3, 5, []uint16{'a', '\x12'}, ""},
		{"", 3, 5, nil, "empty pattern"},
		{"?", 3, 5, nil, "pattern cannot be all '?' (match any byte)"},
		{"????", 3, 5, nil, "pattern cannot be all '?' (match any byte)"},
	}
	for _, c := range cases {
		s, e := NewSearcher(c.inputPattern, c.inputBeforeBytes, c.inputAfterBytes)
		if e == nil && c.expectedErrStr != "" {
			t.Errorf("input: %q, err nil, but expected: %q", c.inputPattern, c.expectedErrStr)
		}
		if e != nil && e.Error() != c.expectedErrStr {
			t.Errorf("input: %q, unexpected err, expected: %q, got: %q", c.inputPattern, c.expectedErrStr, e.Error())
		}
		if s == nil {
			continue
		}
		if !reflect.DeepEqual(s.pattern, c.expectedPattern) {
			t.Errorf("input: %q, unexpected pattern value, expected: %q, got: %q", c.inputPattern, c.expectedPattern, s.pattern)
		}
		if s.showBeforeBytes != c.inputBeforeBytes {
			t.Errorf("input: %q, unexpected showBeforeBytes value, expected: %d, got: %d", c.inputPattern, c.inputBeforeBytes, s.showBeforeBytes)
		}
		if s.showAfterBytes != c.inputAfterBytes {
			t.Errorf("input: %q, unexpected showAfterBytes value, expected: %d, got: %d", c.inputPattern, c.inputAfterBytes, s.showAfterBytes)
		}
		// should pre-allocate needed buffer space
		if cap(s.matchBuffer) != len(s.pattern) {
			t.Errorf("input: %q, unexpected matchBuffer cap() value, expected: %d, got: %d", c.inputPattern, len(c.expectedPattern), cap(s.matchBuffer))
		}
		if len(s.matchBuffer) != 0 {
			t.Errorf("input: %q, unexpected matchBuffer len() value, expected: %d, got: %d", c.inputPattern, 0, len(s.matchBuffer))
		}
	}
}
