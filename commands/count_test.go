package commands

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_CountBytes(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.",
	)
	expected := "35 bytes"
	ioInfo := options.IOInfo{}
	cmdOpts := []string{}
	err := CountBytes(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests when we have enough data to output KB units
func Test_CountBytes_WithUnits(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.",
	)
	expected := "175 bytes\n0.17 KB"
	ioInfo := options.IOInfo{}
	cmdOpts := []string{}
	err := CountBytes(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests that any options trigger a failure as this command accepts no args/opts
func Test_CountBytes_InvalidOpts(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.",
	)
	ioInfo := options.IOInfo{}
	cmdOpts := []string{"bogus"}
	err := CountBytes(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64}, cmdOpts)
	if err == nil {
		t.Errorf("Exected error when any arguments are given to count command. (no supported args)")
	}
}
