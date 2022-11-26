package commands

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_StringsUtf8(t *testing.T) {
	var writer strings.Builder
	// NOTE: split strings by non-ascii bytes, not spaces
	// BUT: we trim leading/trailing whitespace around strings once split out
	reader := strings.NewReader(
		`Hello, world.
新年快樂！
¡Buenos días!
スペインの雨は主に平野に降る。
Deyin, çiləməyin!
The end.`,
	)
	/*
	    0:  Hello, world.
	    E:  新年快樂！
	   1E:  ¡Buenos días!
	   2E:  スペインの雨は主に平野に降る。
	   5C:  Deyin, çiləməyin!
	   71:  The end.
	*/
	expected := "            0:\tHello, world.\n            E:\t新年快樂！\n           1E:\t¡Buenos días!\n           2E:\tスペインの雨は主に平野に降る。\n           5C:\tDeyin, çiləməyin!\n           71:\tThe end."
	ioInfo := options.IOInfo{}
	cmdOpts := []string{}
	err := StringsUtf8(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests that min len arg works
func Test_StringsUtf8_MinMaxLen(t *testing.T) {
	var writer strings.Builder
	// NOTE: split strings by non-ascii bytes, not spaces
	// BUT: we trim leading/trailing whitespace around strings once split out
	// NOTE: extra whitespace before/after two of our captured strings.
	// Should exclude left/right whitespace from the strings, but correctly show the
	// start of the trimmed string. (offset/start at first non-whitespace byte)
	reader := strings.NewReader(
		`Hello, world.
  新年快樂！
¡Buenos días!
   スペインの雨は主に平野に降る。
Deyin, çiləməyin!
The end.`,
	)
	/*
	    E:  新年快樂！
	   74:  The end.
	*/
	expected := "           10:\t新年快樂！\n           76:\tThe end."
	ioInfo := options.IOInfo{}
	cmdOpts := []string{"5", "10"}
	err := StringsUtf8(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_StringsUtf8_InvalidBytes(t *testing.T) {
	var writer strings.Builder
	// NOTE: before/after the foreign chars we have what starts off looking like utf8 but doesn't
	// have all the bytes (b1110 start indicating 3 bytes but only one byte following with high bits 0x01
	// for continuation.)

	// Valid utf8 char that is 4 bytes:
	// U+10348
	// 11110000 10010000 10001101 10001000	F0 90 8D 88

	// Include various invalid forms of this unicode rune
	// ex: leave off last byte, leave off first byte etc

	reader := strings.NewReader(
		"\x00\x01\x02\x03\xF0\x90\x8D新年\x90\x8D\x88快樂\x88\x31The End.",
	)
	// Should get the 3 sections of valid utf8 chars as 3 separate utf8 string outputs:
	expected := "            7:\t新年\n           10:\t快樂\n           17:\t1The End."
	ioInfo := options.IOInfo{}
	cmdOpts := []string{}
	err := StringsUtf8(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
