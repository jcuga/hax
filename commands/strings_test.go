package commands

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_Strings(t *testing.T) {
	var writer strings.Builder
	// NOTE: split strings by non-ascii bytes, not spaces
	// BUT: we trim leading/trailing whitespace around strings once split out
	reader := strings.NewReader(
		"Some string\x01\x02\x03 and another one              \x04\x05\x06Okay!\x0aDone.",
	)
	/*
	    0:	Some string
	    F:	and another one
	   2F:	Okay!
	   35:	Done.
	*/
	expected := "            0:\tSome string\n            F:\tand another one\n           2F:\tOkay!\n           35:\tDone."
	ioInfo := options.IOInfo{}
	cmdOpts := []string{}
	err := Strings(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests that min len arg works
func Test_Strings_MinLen(t *testing.T) {
	var writer strings.Builder
	// NOTE: split strings by non-ascii bytes, not spaces
	// BUT: we trim leading/trailing whitespace around strings once split out
	reader := strings.NewReader(
		"Some string\x01\x02\x03 and another one              \x04\x05\x06Okay!\x0aDone.",
	)
	/*
	   0:	Some string
	   F:	and another one
	*/
	expected := "            0:\tSome string\n            F:\tand another one"
	ioInfo := options.IOInfo{}
	cmdOpts := []string{"11"}
	err := Strings(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests that min and max len arg works
func Test_Strings_MinMaxLen(t *testing.T) {
	var writer strings.Builder
	// NOTE: split strings by non-ascii bytes, not spaces
	// BUT: we trim leading/trailing whitespace around strings once split out
	reader := strings.NewReader(
		"Some string\x01\x02\x03 and another one              \x04\x05\x06Okay!\x0aDone.",
	)
	/*
	   0:	Some string
	*/
	expected := "            0:\tSome string"
	ioInfo := options.IOInfo{}
	cmdOpts := []string{"11", "13"}
	err := Strings(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
