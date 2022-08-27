package output

import (
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_outputHexString(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// NOTE: expect a newline (0A) at the end when not a pipe output
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\\x20\\x69\\x6E\\x20\\x53\\x70\\x61\\x69\\x6E\\x20\\x66\\x61\\x6C\\x6C\\x73\\x20\\x6D\\x61\\x69\\x6E\\x6C\\x79\\x20\\x69\\x6E\\x20\\x74\\x68\\x65\\x20\\x70\\x6C\\x61\\x69\\x6E\\x73\\x2E\x0A"
	outputHexString(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{Limit: math.MaxInt64})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q",expected, result)
	}

	// NOTE: expect no injected newline at end when isPipe is true
	writer.Reset()
	reader = strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	outputHexString(&writer, input.NewFixedLengthBufferedReader(reader), true,
		options.Options{Limit: math.MaxInt64})
	result = writer.String()
	if result != expected[:len(expected)-1] {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q", expected[:len(expected)-1], result)
	}
}

// NOTE: not testing Offset as that's enforced at the input stage, not the output stage!
func Test_outputHexString_WithLimit(t *testing.T) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	// NOTE: expect a newline (0A) at the end when not a pipe output
	// Expect escaped hex of "The rain". Plus newline.
	expected := "\\x54\\x68\\x65\\x20\\x72\\x61\\x69\\x6E\n"
	outputHexString(&writer, input.NewFixedLengthBufferedReader(reader), false,
		options.Options{Limit: 8})
	result := writer.String()
	if result != expected {
		t.Errorf("Unexpected output.\nExpected:\n%q\n\ngot:\n%q",expected, result)
	}
}

func Benchmark_outputHexString(b *testing.B) {
	var writer strings.Builder
	reader := strings.NewReader(
		"The rain in Spain falls mainly in the plains.",
	)
	for i := 0 ; i < b.N ; i++ {
		outputHexString(&writer, input.NewFixedLengthBufferedReader(reader), true,
			options.Options{Limit: math.MaxInt64})
	}
}
