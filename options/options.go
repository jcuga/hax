package options

import (
	"fmt"
	"math"
	"strings"

	"github.com/jcuga/hax/eval"
)

type IOMode int

const (
	Raw IOMode = iota
	Hex
	HexString // Escaped hex that can be used in a string literal, ex: "\xAB\xCD\xEF"
	HexList   // List of hex bytes that can be used in an array literal: "0xAB, 0xCD, 0xEF"
	Base64
	Display
	Strings
)

type DisplayOptions struct {
	Width          int
	SubWidth       int // add space after every SubWidth bytes
	PageSize       int
	Pretty         bool
	Quiet          bool
	HideZerosBytes bool
	OmitZeroPages  bool
	MinStringLen   int // min len of strings to include in strings output
	MaxStringLen   int
}

type Options struct {
	Filename   string
	InputData  string
	InputMode  IOMode
	OutputMode IOMode
	Offset     int64
	Limit      int64
	Display    DisplayOptions
	// Yes is whether to auto-answer y/yes to any prompts
	Yes bool
}

func parseInputMode(mode string) (IOMode, error) {
	normMode := strings.ToLower(mode)
	switch normMode {
	case "raw", "r":
		return Raw, nil
	case "hex", "h":
		return Hex, nil
	case "hex-string", "hex-str", "hexstr", "hs", "str", "s":
		return HexString, nil
	case "hex-list", "hexlist", "hl", "list", "l":
		return HexList, nil
	case "base64", "b64", "b":
		return Base64, nil
	// NOTE: not valid input modes: Display, Strings
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
	case "hex-string", "hex-str", "hexstr", "hs":
		return HexString, nil
	case "hex-list", "hexlist", "hl", "list", "l":
		return HexList, nil
	case "base64", "b64", "b":
		return Base64, nil
	case "display", "d":
		return Display, nil
	case "strings", "string", "str", "strs", "s":
		return Strings, nil
	default:
		return -1, fmt.Errorf("Not a valid output mode: %q.", mode)
	}
}

// TODO: any error text here needs arg names to update those in main if they've changed during development!
func New(inFilename, inputStr, inMode, outMode, offset, limit, colWidth, colSubWidth,
	pageSize string, alwaysPretty, quiet, yes, hideZeros, omitZeroPages bool,
	minStringLen, maxStringLen int) (Options, error) {

	opts := Options{
		Filename:  inFilename,
		InputData: inputStr,
		Display: DisplayOptions{
			Pretty:         alwaysPretty,
			Quiet:          quiet,
			HideZerosBytes: hideZeros,
			OmitZeroPages:  omitZeroPages,
			MinStringLen:   minStringLen,
			MaxStringLen:   maxStringLen,
		},
		Yes: yes,
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

	opts.Offset = 0
	if len(offset) > 0 {
		if parsedOffset, err := eval.EvalExpression(offset); err == nil {
			if parsedOffset < 0 {
				return opts, fmt.Errorf(
					"Invalid --offset/-n value %q, must be >= 0 ", offset)
			}
			opts.Offset = parsedOffset
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --offset/-n value %q, error: %v", offset, err)
		}
	}

	opts.Limit = math.MaxInt64
	if len(limit) > 0 {
		if parsedLimit, err := eval.EvalExpression(limit); err == nil {
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
	}

	if colWidth == "" {
		if opts.OutputMode == Display {
			opts.Display.Width = 16
		} else {
			// NOTE: widht of 0 implies infinite/no width limit on output
			opts.Display.Width = 0
		}
	} else {
		if parsedWidth, err := eval.EvalExpression(colWidth); err == nil {
			if parsedWidth < 1 || parsedWidth > 1024 {
				return opts, fmt.Errorf(
					"Invalid --width/-w value %q, must be 1-1024 ", colWidth)
			}
			opts.Display.Width = int(parsedWidth)
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --width/-w value %q, error: %v", colWidth, err)
		}
	}

	if colSubWidth == "" {
		opts.Display.SubWidth = 0
	} else {
		if parsedSubWidth, err := eval.EvalExpression(colSubWidth); err == nil {
			if parsedSubWidth < 0 || (opts.Display.Width > 0 && parsedSubWidth > int64(opts.Display.Width)) {
				return opts, fmt.Errorf(
					"Invalid --sub-width/-ww value %q, must be 0 to --width (%d) ", colSubWidth, opts.Display.Width)
			}
			opts.Display.SubWidth = int(parsedSubWidth)
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --sub-width/-ww value %q, error: %v", colSubWidth, err)
		}
	}

	if parsedPage, err := eval.EvalExpression(pageSize); err == nil {
		if parsedPage < 0 {
			return opts, fmt.Errorf(
				"Invalid --page/-p value %q, must be >= 0", colWidth)
		}

		opts.Display.PageSize = int(parsedPage)
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --page/-p value %q, error: %v", pageSize, err)
	}

	if opts.Display.MaxStringLen < 0 {
		opts.Display.MaxStringLen = math.MaxInt32
	}

	if opts.Display.MaxStringLen < opts.Display.MinStringLen {
		return opts, fmt.Errorf("Invalid --max-str: %d, must be > --min-str: %d",
			opts.Display.MaxStringLen, opts.Display.MinStringLen)
	}

	return opts, nil
}
