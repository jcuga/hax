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
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\\x20\\x69\\x6E\\x20\\x53\\x70\\x61\\x69\\x6E\\x20\\x66\\x61\\x6C\\x6C\\x73\\x20\\x6D\\x61\\x69\\x6E\\x6C\\x79\\x20\\x69\\x6E\\x20\\x74\\x68\\x65\\x20\\x70\\x6C\\x61\\x69\\x6E\\x73\\x2E"
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

// NOTE: not testing Offset as that's enforced at the input stage, not the output stage!
func Test_outputHexStringOrList_HexString_WithLimit(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E"
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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
\x61\x69\x6E\x73\x2E`
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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
	expected := "0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E, 0x20, 0x69, 0x6E, 0x20, 0x53, 0x70, 0x61, 0x69, 0x6E, 0x20, 0x66, 0x61, 0x6C, 0x6C, 0x73, 0x20, 0x6D, 0x61, 0x69, 0x6E, 0x6C, 0x79, 0x20, 0x69, 0x6E, 0x20, 0x74, 0x68, 0x65, 0x20, 0x70, 0x6C, 0x61, 0x69, 0x6E, 0x73, 0x2E"
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexList, Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

// NOTE: not testing Offset as that's enforced at the input stage, not the output stage!
func Test_outputHexStringOrList_HexList_WithLimit(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	expected := "0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E"
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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
0x61, 0x69, 0x6E, 0x73, 0x2E`
	ioInfo := options.IOInfo{}
	outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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
	expected := "\n\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\\x20\\x69\\x6E\\x20\\x53\\x70\\x61\\x69\\x6E\\x20\\x66\\x61\\x6C\\x6C\\x73\\x20\\x6D\\x61\\x69\\x6E\\x6C\\x79\\x20\\x69\\x6E\\x20\\x74\\x68\\x65\\x20\\x70\\x6C\\x61\\x69\\x6E\\x73\\x2E\x0A"
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		reader := strings.NewReader(
			"The rain in Spain falls mainly in the plains.",
		)
		ioInfo := options.IOInfo{}
		outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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
	expected := "\n0x54, 0x68, 0x65, 0x20, 0x72, 0x61, 0x69, 0x6E, 0x20, 0x69, 0x6E, 0x20, 0x53, 0x70, 0x61, 0x69, 0x6E, 0x20, 0x66, 0x61, 0x6C, 0x6C, 0x73, 0x20, 0x6D, 0x61, 0x69, 0x6E, 0x6C, 0x79, 0x20, 0x69, 0x6E, 0x20, 0x74, 0x68, 0x65, 0x20, 0x70, 0x6C, 0x61, 0x69, 0x6E, 0x73, 0x2E\x0A"
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		reader := strings.NewReader(
			"The rain in Spain falls mainly in the plains.",
		)
		ioInfo := options.IOInfo{}
		outputHexStringOrList(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
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

func Test_outputHexAscii(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"Counting \x01\x02\x03 and a \x04\x05\x06 seven!\x0aDone.",
	)
	expected := "Counting \\x01\\x02\\x03 and a \\x04\\x05\\x06 seven!\\nDone."
	ioInfo := options.IOInfo{}
	outputHexAscii(&writer, input.NewFixedLengthBufferedReader(reader), ioInfo,
		options.Options{OutputMode: options.HexString, Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}
