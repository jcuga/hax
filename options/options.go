package options

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type IOMode int

const (
	Raw IOMode = iota
	Hex
	Base64
	Display
)

type DisplayOptions struct {
	Width    int
	PageSize int
	Pretty   bool
	NoAscii  bool
}

type Options struct {
	Filename   string
	InputData  string
	InputMode  IOMode
	OutputMode IOMode
	Offset     int64
	Limit      int64
	Display    DisplayOptions
}

func parseInputMode(mode string) (IOMode, error) {
	normMode := strings.ToLower(mode)
	switch normMode {
	case "raw", "r":
		return Raw, nil
	case "hex", "h":
		return Hex, nil
	case "base64", "b64", "b":
		return Base64, nil
	// NOTE: not valid input modes: Display
	default:
		return -1, fmt.Errorf("Not a valid input mode: %q.", mode)
	}
}

func parseOutputMode(mode string) (IOMode, error) {
	normMode := strings.ToLower(mode)
	switch normMode {
	case "raw", "r":
		return Raw, nil
	case "hex", "h":
		return Hex, nil
	case "base64", "b64", "b":
		return Base64, nil
	case "display", "d":
		return Display, nil
	default:
		return -1, fmt.Errorf("Not a valid output mode: %q.", mode)
	}
}

func New(inFilename, inputStr, inMode, outMode, offset, limit, colWidth,
	pageSize string, alwaysPretty, quiet bool) (Options, error) {

	opts := Options{
		Filename:  inFilename,
		InputData: inputStr,
		Display: DisplayOptions{
			Pretty:  alwaysPretty,
			NoAscii: quiet,
		},
	}

	if len(inMode) > 0 {
		if mode, err := parseInputMode(inMode); err == nil {
			opts.InputMode = mode
		} else {
			return opts, fmt.Errorf("Invalid --input/-i value. %v ", err)
		}
	} else {
		// default to Hex when --str input, otherwise raw for --file/stdin
		if len(inputStr) > 0 {
			opts.InputMode = Hex
		} else {
			opts.InputMode = Raw
		}
	}

	if len(outMode) > 0 {
		if mode, err := parseOutputMode(outMode); err == nil {
			opts.OutputMode = mode
		} else {
			return opts, fmt.Errorf("Invalid --output/-o value. %v ", err)
		}
	} else {
		// display hex editor output by default
		opts.OutputMode = Display
	}

	if parsedOffset, err := parseHexOrDec(offset); err == nil {
		if parsedOffset < 0 {
			return opts, fmt.Errorf(
				"Invalid --offset/-n value %q, must be >= 0 ", offset)
		}
		opts.Offset = parsedOffset
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --offset/-n value %q, error: %v", offset, err)
	}

	if parsedLimit, err := parseHexOrDec(limit); err == nil {
		if parsedLimit < 0 {
			return opts, fmt.Errorf(
				"Invalid --limit/-l value %q, must be >= 0 ", limit)
		}
		if parsedLimit == 0 {
			opts.Limit = math.MaxInt64
		} else {
			opts.Limit = parsedLimit
		}
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --limit/-l value %q, error: %v", limit, err)
	}

	if parsedWidth, err := parseHexOrDec(colWidth); err == nil {
		if parsedWidth < 1 || parsedWidth > 1024 {
			return opts, fmt.Errorf(
				"Invalid --width/-w value %q, must be 1-1024 ", colWidth)
		}
		opts.Display.Width = int(parsedWidth)
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --width/-w value %q, error: %v", colWidth, err)
	}

	if parsedPage, err := parseHexOrDec(pageSize); err == nil {
		if parsedPage < 0 {
			return opts, fmt.Errorf(
				"Invalid --page/-p value %q, must be >= 0 ", colWidth)
		}
		opts.Display.PageSize = int(parsedPage)
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --page/-p value %q, error: %v", pageSize, err)
	}

	return opts, nil
}

func parseHexOrDec(input string) (int64, error) {
	if len(input) == 0 {
		return 0, nil
	}
	// If sarts with "0", "x", or "0x", "\x" OR has a-fA-F, then interpret as hex
	// also interpret as hex if has spaces between nubmers which would be if one
	// copy-pasted a value from a previous run's output (ex: "AA BB CC")
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "0") || strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "x") ||
		strings.HasPrefix(input, "\\x") || strings.ContainsAny(input, "abcdef ") {
		// assume hex
		// trim off leading "0x" or "x" if found (could be just 0 in front)
		xIndex := strings.Index(input, "x")
		if xIndex != -1 {
			input = input[xIndex+1:]
		}
		// remove any spaces between bytes ("AA BB" --> "AABB")
		input = strings.Replace(input, " ", "", -1)
		return strconv.ParseInt(input, 16, 64)
	}
	return strconv.ParseInt(input, 10, 64)
}
