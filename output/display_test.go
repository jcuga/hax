package output

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Test_Output_displayHex(t *testing.T) {
	var writer strings.Builder
	// Below is same as: b'This is only a test.\x00\x00\x00Countdown in:\n\t\x08\x07\x06\x05\x04\x03\x02\x01\x00--LIFTOFF!\x00\x00'
	original := "54686973206973206f6e6c79206120746573742e000000436f756e74646f776e20696e3a0a090807060504030201002d2d4c4946544f4646210000"
	expected := `
                0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F 
            0: 54 68 69 73 20 69 73 20 6F 6E 6C 79 20 61 20 74 
                T  h  i  s     i  s     o  n  l  y     a     t 
           10: 65 73 74 2E 00 00 00 43 6F 75 6E 74 64 6F 77 6E 
                e  s  t  .           C  o  u  n  t  d  o  w  n 
           20: 20 69 6E 3A 0A 09 08 07 06 05 04 03 02 01 00 2D 
                   i  n  : \n \t                             - 
           30: 2D 4C 49 46 54 4F 46 46 21 00 00 
                -  L  I  F  T  O  F  F  !       

` // Note the newlines at end of this str.
	opts1 := options.Options{
		InputMode:  options.Hex,
		OutputMode: options.Display,
		InputData:  original,
		Limit:      math.MaxInt64,
		Display:    options.DisplayOptions{Width: 16},
	}
	reader, closer, err := input.GetInput(opts1)
	if err != nil {
		t.Fatalf("Failed to create input reader, error: %v", err)
	}
	if closer != nil {
		t.Errorf("Expect nil closer when GetInput on string data, got: %v", closer)
	}
	Output(&writer, reader, true, opts1, options.NoCommand, []string{}) // NOTE: isPipe:true for plain text instead of terminal colors
	result := writer.String()
	if result != expected {
		t.Fatalf("Unexpected hex display output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Test_Output_displayHex_HideZeroBytes(t *testing.T) {
	var writer strings.Builder
	// Below is same as: b'This is only a test.\x00\x00\x00Countdown in:\n\t\x08\x07\x06\x05\x04\x03\x02\x01\x00--LIFTOFF!\x00\x00'
	original := "54686973206973206f6e6c79206120746573742e000000436f756e74646f776e20696e3a0a090807060504030201002d2d4c4946544f4646210000"
	expected := `
                0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F 
            0: 54 68 69 73 20 69 73 20 6F 6E 6C 79 20 61 20 74 
                T  h  i  s     i  s     o  n  l  y     a     t 
           10: 65 73 74 2E          43 6F 75 6E 74 64 6F 77 6E 
                e  s  t  .           C  o  u  n  t  d  o  w  n 
           20: 20 69 6E 3A 0A 09 08 07 06 05 04 03 02 01    2D 
                   i  n  : \n \t                             - 
           30: 2D 4C 49 46 54 4F 46 46 21       
                -  L  I  F  T  O  F  F  !       

` // Note the newlines at end of this str.
	opts1 := options.Options{
		InputMode:  options.Hex,
		OutputMode: options.Display,
		InputData:  original,
		Limit:      math.MaxInt64,
		Display:    options.DisplayOptions{Width: 16, HideZerosBytes: true},
	}
	reader, closer, err := input.GetInput(opts1)
	if err != nil {
		t.Fatalf("Failed to create input reader, error: %v", err)
	}
	if closer != nil {
		t.Errorf("Expect nil closer when GetInput on string data, got: %v", closer)

	}
	Output(&writer, reader, true, opts1, options.NoCommand, []string{}) // NOTE: isPipe:true for plain text instead of terminal colors
	result := writer.String()
	if result != expected {
		t.Fatalf("Unexpected hex display output.\nExpected:\n%q\n\ngot:\n%q", expected, result)
	}
}

func Benchmark_output_displayHex(b *testing.B) {
	// Below is same as: b'This is only a test.\x00\x00\x00Countdown in:\n\t\x08\x07\x06\x05\x04\x03\x02\x01\x00--LIFTOFF!\x00\x00'
	original := "54686973206973206f6e6c79206120746573742e000000436f756e74646f776e20696e3a0a090807060504030201002d2d4c4946544f4646210000"
	expected := `
                0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F 
            0: 54 68 69 73 20 69 73 20 6F 6E 6C 79 20 61 20 74 
                T  h  i  s     i  s     o  n  l  y     a     t 
           10: 65 73 74 2E 00 00 00 43 6F 75 6E 74 64 6F 77 6E 
                e  s  t  .           C  o  u  n  t  d  o  w  n 
           20: 20 69 6E 3A 0A 09 08 07 06 05 04 03 02 01 00 2D 
                   i  n  : \n \t                             - 
           30: 2D 4C 49 46 54 4F 46 46 21 00 00 
                -  L  I  F  T  O  F  F  !       

` // Note the newlines at end of this str.
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		opts1 := options.Options{
			InputMode:  options.Hex,
			OutputMode: options.Display,
			InputData:  original,
			Limit:      math.MaxInt64,
			Display:    options.DisplayOptions{Width: 16},
		}
		reader, closer, err := input.GetInput(opts1)
		if err != nil {
			panic(fmt.Sprintf("Failed to create input reader, error: %v", err))
		}
		if closer != nil {
			panic(fmt.Sprintf("Expect nil closer when GetInput on string data, got: %v", closer))
		}
		Output(&writer, reader, true, opts1, options.NoCommand, []string{}) // NOTE: isPipe:true for plain text instead of terminal colors
		result := writer.String()
		if result != expected {
			panic(fmt.Sprintf("Unexpected hex display output.\nExpected:\n%q\n\ngot:\n%q", expected, result))
		}
	}
}

func Benchmark_output_displayHex_HideZeroBytes(b *testing.B) {
	// Below is same as: b'This is only a test.\x00\x00\x00Countdown in:\n\t\x08\x07\x06\x05\x04\x03\x02\x01\x00--LIFTOFF!\x00\x00'
	original := "54686973206973206f6e6c79206120746573742e000000436f756e74646f776e20696e3a0a090807060504030201002d2d4c4946544f4646210000"
	expected := `
                0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F 
            0: 54 68 69 73 20 69 73 20 6F 6E 6C 79 20 61 20 74 
                T  h  i  s     i  s     o  n  l  y     a     t 
           10: 65 73 74 2E          43 6F 75 6E 74 64 6F 77 6E 
                e  s  t  .           C  o  u  n  t  d  o  w  n 
           20: 20 69 6E 3A 0A 09 08 07 06 05 04 03 02 01    2D 
                   i  n  : \n \t                             - 
           30: 2D 4C 49 46 54 4F 46 46 21       
                -  L  I  F  T  O  F  F  !       

` // Note the newlines at end of this str.
	for i := 0; i < b.N; i++ {
		var writer strings.Builder
		opts1 := options.Options{
			InputMode:  options.Hex,
			OutputMode: options.Display,
			InputData:  original,
			Limit:      math.MaxInt64,
			Display:    options.DisplayOptions{Width: 16, HideZerosBytes: true},
		}
		reader, closer, err := input.GetInput(opts1)
		if err != nil {
			panic(fmt.Sprintf("Failed to create input reader, error: %v", err))
		}
		if closer != nil {
			panic(fmt.Sprintf("Expect nil closer when GetInput on string data, got: %v", closer))

		}
		Output(&writer, reader, true, opts1, options.NoCommand, []string{}) // NOTE: isPipe:true for plain text instead of terminal colors
		result := writer.String()
		if result != expected {
			panic(fmt.Sprintf("Unexpected hex display output.\nExpected:\n%q\n\ngot:\n%q", expected, result))
		}
	}
}
