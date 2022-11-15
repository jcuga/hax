package output

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

// Test able to output hexlist data then parse it back in to get original value back.
// Tests that data does not get garbled either in or out.
func Test_Output_HexListOutputThenInputToRaw(t *testing.T) {
	var writer strings.Builder
	original := "I'm a little tea pot, short and stout."
	expected := `0x49, 0x27, 0x6D, 0x20, 0x61, 0x20, 0x6C, 0x69, 
0x74, 0x74, 0x6C, 0x65, 0x20, 0x74, 0x65, 0x61, 
0x20, 0x70, 0x6F, 0x74, 0x2C, 0x20, 0x73, 0x68, 
0x6F, 0x72, 0x74, 0x20, 0x61, 0x6E, 0x64, 0x20, 
0x73, 0x74, 0x6F, 0x75, 0x74, 0x2E
` // Note the newline at end of this str.
	opts1 := options.Options{
		InputMode:  options.Raw,
		OutputMode: options.HexList,
		InputData:  original, // NOTE: setting str input (versus file or piped)
		Limit:      math.MaxInt64,
		Display:    options.DisplayOptions{Width: 8},
	}
	// NOTE: use GetInput to get wrapped reader with proper "mode" based on opts.InputMode.
	// One can't simply set the input mode on opts and pass it to Output().
	// Output does not modify/wrap the input reader, the GetInput func does.
	reader, closer, isStdin, err := input.GetInput(opts1)
	if err != nil {
		t.Fatalf("Failed to create input reader, error: %v", err)
	}
	if closer != nil {
		t.Errorf("Expect nil closer when GetInput on string data, got: %v", closer)

	}
	Output(&writer, reader, false, isStdin, opts1, options.NoCommand, []string{})
	result := writer.String()
	if result != expected {
		t.Fatalf("Unexpected hexlist output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}

	// Now feed the hexlist data back in and ensure we get orignal value back.
	writer.Reset()
	opts2 := options.Options{
		InputMode:  options.HexList,
		OutputMode: options.Raw,
		InputData:  result, // NOTE: feeding last output as this input
		Limit:      math.MaxInt64,
		Display:    options.DisplayOptions{Width: 8},
	}
	reader, closer, isStdin, err = input.GetInput(opts2)
	if err != nil {
		t.Fatalf("Failed to create input reader, error: %v", err)
	}
	if closer != nil {
		t.Errorf("Expect nil closer when GetInput on string data, got: %v", closer)

	}
	Output(&writer, input.NewFixedLengthBufferedReader(reader), true, isStdin, // NOTE: isPipe true will cause no
		// extra preceeding/ending newlines to be added.
		options.Options{
			InputMode:  options.HexList,
			OutputMode: options.Raw,
			Limit:      math.MaxInt64,
		}, options.NoCommand, []string{})
	result = writer.String()
	if result != original {
		t.Fatalf("Unexpected raw output.\nExpected:\n%q\n\ngot:\n%q", original, result)
	}
}
