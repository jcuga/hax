package output

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_outputHexStringOrList_HexString(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// NOTE: expect a newline (0A) at the end when not a pipe output
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\\x20\\x69\\x6E\\x20\\x53\\x70\\x61\\x69\\x6E\\x20\\x66\\x61\\x6C\\x6C\\x73\\x20\\x6D\\x61\\x69\\x6E\\x6C\\x79\\x20\\x69\\x6E\\x20\\x74\\x68\\x65\\x20\\x70\\x6C\\x61\\x69\\x6E\\x73\\x2E\x0A"
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}

	// NOTE: expect no injected newline at end when isPipe is true
	writer.Reset()
	reader = strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), true,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64})
	result = writer.String()
	if result != expected[:len(expected)-1] {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected[:len(expected)-1], result)
	}
}

// NOTE: not testing Offset as that's enforced at the input stage, not the output stage!
func Test_outputHexStringOrList_HexString_WithLimit(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// Expect escaped hex of "The rain". Plus newline.
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\n"
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{OutputMode: options.HexString, Limit: 8})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Test_outputHexStringOrList_HexString_WithWidth(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	expected := `\x54\x68\x65\x20\x72\x61\x69\x6E\x20\x69
\x6E\x20\x53\x70\x61\x69\x6E\x20\x66\x61
\x6C\x6C\x73\x20\x6D\x61\x69\x6E\x6C\x79
\x20\x69\x6E\x20\x74\x68\x65\x20\x70\x6C
\x61\x69\x6E\x73\x2E
` // Note the newline at end of this str.
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{
			OutputMode: options.HexString,
			Limit:      math.MaxInt64,
			Display:    options.DisplayOptions{Width: 10},
		})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Test_outputHexStringOrList_HexList(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// NOTE: expect a newline (0A) at the end when not a pipe output
	expected := "0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E, 0x20, 0x69, 0x6E, 0x20, 0x53, 0x70, 0x61, 0x69, 0x6E, 0x20, 0x66, 0x61, 0x6C, 0x6C, 0x73, 0x20, 0x6D, 0x61, 0x69, 0x6E, 0x6C, 0x79, 0x20, 0x69, 0x6E, 0x20, 0x74, 0x68, 0x65, 0x20, 0x70, 0x6C, 0x61, 0x69, 0x6E, 0x73, 0x2E\x0A"
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{OutputMode: options.HexList, Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}

	// NOTE: expect no injected newline at end when isPipe is true
	writer.Reset()
	reader = strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), true,
		options.Options{OutputMode: options.HexList, Limit: math.MaxInt64})
	result = writer.String()
	if result != expected[:len(expected)-1] {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected[:len(expected)-1], result)
	}
}

// NOTE: not testing Offset as that's enforced at the input stage, not the output stage!
func Test_outputHexStringOrList_HexList_WithLimit(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// Expect escaped hex of "The rain". Plus newline.
	expected := "0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E\n"
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{OutputMode: options.HexList, Limit: 8})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Test_outputHexStringOrList_HexList_WithWidth(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	expected := `0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E, 0x20, 0x69, 
0x6E, 0x20, 0x53, 0x70, 0x61, 0x69, 0x6E, 0x20, 0x66, 0x61, 
0x6C, 0x6C, 0x73, 0x20, 0x6D, 0x61, 0x69, 0x6E, 0x6C, 0x79, 
0x20, 0x69, 0x6E, 0x20, 0x74, 0x68, 0x65, 0x20, 0x70, 0x6C, 
0x61, 0x69, 0x6E, 0x73, 0x2E
` // Note the newline at end of this str.
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{
			OutputMode: options.HexList,
			Limit:      math.MaxInt64,
			Display:    options.DisplayOptions{Width: 10},
		})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Benchmark_outputHexStringOrList_HexString(b *testing.B) {
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\\x20\\x69\\x6E\\x20\\x53\\x70\\x61\\x69\\x6E\\x20\\x66\\x61\\x6C\\x6C\\x73\\x20\\x6D\\x61\\x69\\x6E\\x6C\\x79\\x20\\x69\\x6E\\x20\\x74\\x68\\x65\\x20\\x70\\x6C\\x61\\x69\\x6E\\x73\\x2E\x0A"
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		reader := strings.NewReader(
			"The rain in Spain falls mainly in the plains.",
		)
		outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
			options.Options{OutputMode: options.HexString, Limit: math.MaxInt64})
		result := writer.String()
		// Ensure we're actually creating expected output.
		// Avoids any scenario where trivial output is happening when it should
		// be doing more.
		if result != expected {
			panic(result)
		}
	}
}

func Benchmark_outputHexStringOrList_HexList(b *testing.B) {
	expected := "0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E, 0x20, 0x69, 0x6E, 0x20, 0x53, 0x70, 0x61, 0x69, 0x6E, 0x20, 0x66, 0x61, 0x6C, 0x6C, 0x73, 0x20, 0x6D, 0x61, 0x69, 0x6E, 0x6C, 0x79, 0x20, 0x69, 0x6E, 0x20, 0x74, 0x68, 0x65, 0x20, 0x70, 0x6C, 0x61, 0x69, 0x6E, 0x73, 0x2E\x0A"
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		reader := strings.NewReader(
			"The rain in Spain falls mainly in the plains.",
		)
		outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), false,
			options.Options{OutputMode: options.HexList, Limit: math.MaxInt64})
		result := writer.String()
		// Ensure we're actually creating expected output.
		// Avoids any scenario where trivial output is happening when it should
		// be doing more.
		if result != expected {
			panic(result)
		}
	}
}
