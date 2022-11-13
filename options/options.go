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
	// TODO: these will get removed in favor of cmd pattern
	MinStringLen int // min len of strings to include in strings output
	MaxStringLen int
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

// RawOptions are pre-parsed, pre-validated version of options.
// Used to conveniently pass around all the options instead of N func args.
// Some numeric options are input as string to allow expression evaluation and various
// representations (excaped hex, binary, etc).
type RawOptions struct {
	Filename   string
	InputData  string
	InputMode  string
	OutputMode string
	Offset     string
	Limit      string
	Display    RawDisplayOptions
	// Yes is whether to auto-answer y/yes to any prompts
	Yes bool
}

// RawDisplayOptions are pre-parsed, pre-validated options.\
// Proivded as an alternative to passing N fucn args.
type RawDisplayOptions struct {
	Width          string
	SubWidth       string // add space after every SubWidth bytes
	PageSize       string
	Pretty         bool
	Quiet          bool
	HideZerosBytes bool
	OmitZeroPages  bool
	// TODO: these will get removed in favor of cmd pattern
	MinStringLen int // min len of strings to include in strings output
	MaxStringLen int
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
func New(rawOpts RawOptions) (Options, error) {

	opts := Options{
		Filename:  rawOpts.Filename,
		InputData: rawOpts.InputData,
		Display: DisplayOptions{
			Pretty:         rawOpts.Display.Pretty,
			Quiet:          rawOpts.Display.Quiet,
			HideZerosBytes: rawOpts.Display.HideZerosBytes,
			OmitZeroPages:  rawOpts.Display.OmitZeroPages,
			MinStringLen:   rawOpts.Display.MinStringLen,
			MaxStringLen:   rawOpts.Display.MaxStringLen,
		},
		Yes: rawOpts.Yes,
	}

	if len(rawOpts.InputMode) > 0 {
		if mode, err := parseInputMode(rawOpts.InputMode); err == nil {
			opts.InputMode = mode
		} else {
			// TODO: update these to not use the --name of arg? Higher level name...
			return opts, fmt.Errorf("Invalid --input/-i value. %v ", err)
		}
	} else {
		// default to Hex when --str input, otherwise raw for --file/stdin
		if len(rawOpts.InputData) > 0 {
			opts.InputMode = Hex
		} else {
			opts.InputMode = Raw
		}
	}

	if len(rawOpts.OutputMode) > 0 {
		if mode, err := parseOutputMode(rawOpts.OutputMode); err == nil {
			opts.OutputMode = mode
		} else {
			return opts, fmt.Errorf("Invalid --output/-o value. %v ", err)
		}
	} else {
		// display hex editor output by default
		opts.OutputMode = Display
	}

	opts.Offset = 0
	if len(rawOpts.Offset) > 0 {
		if parsedOffset, err := eval.EvalExpression(rawOpts.Offset); err == nil {
			if parsedOffset < 0 {
				return opts, fmt.Errorf(
					"Invalid --offset/-n value %q, must be >= 0 ", rawOpts.Offset)
			}
			opts.Offset = parsedOffset
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --offset/-n value %q, error: %v", rawOpts.Offset, err)
		}
	}

	opts.Limit = math.MaxInt64
	if len(rawOpts.Limit) > 0 {
		if parsedLimit, err := eval.EvalExpression(rawOpts.Limit); err == nil {
			if parsedLimit < 0 {
				return opts, fmt.Errorf(
					"Invalid --limit/-l value %q, must be >= 0 ", rawOpts.Limit)
			}
			if parsedLimit == 0 {
				opts.Limit = math.MaxInt64
			} else {
				opts.Limit = parsedLimit
			}
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --limit/-l value %q, error: %v", rawOpts.Limit, err)
		}
	}

	if rawOpts.Display.Width == "" {
		if opts.OutputMode == Display {
			opts.Display.Width = 16
		} else {
			// NOTE: widht of 0 implies infinite/no width limit on output
			opts.Display.Width = 0
		}
	} else {
		if parsedWidth, err := eval.EvalExpression(rawOpts.Display.Width); err == nil {
			if parsedWidth < 1 || parsedWidth > 1024 {
				return opts, fmt.Errorf(
					"Invalid --width/-w value %q, must be 1-1024 ", rawOpts.Display.Width)
			}
			opts.Display.Width = int(parsedWidth)
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --width/-w value %q, error: %v", rawOpts.Display.Width, err)
		}
	}

	if rawOpts.Display.SubWidth == "" {
		opts.Display.SubWidth = 0
	} else {
		if parsedSubWidth, err := eval.EvalExpression(rawOpts.Display.SubWidth); err == nil {
			if parsedSubWidth < 0 || (opts.Display.Width > 0 && parsedSubWidth > int64(opts.Display.Width)) {
				return opts, fmt.Errorf(
					"Invalid --sub-width/-ww value %q, must be 0 to --width (%d) ", rawOpts.Display.SubWidth, opts.Display.Width)
			}
			opts.Display.SubWidth = int(parsedSubWidth)
		} else {
			return opts, fmt.Errorf(
				"Failed to parse --sub-width/-ww value %q, error: %v", rawOpts.Display.SubWidth, err)
		}
	}

	if parsedPage, err := eval.EvalExpression(rawOpts.Display.PageSize); err == nil {
		if parsedPage < 0 {
			return opts, fmt.Errorf(
				"Invalid --page/-p value %q, must be >= 0", rawOpts.Display.PageSize)
		}

		opts.Display.PageSize = int(parsedPage)
	} else {
		return opts, fmt.Errorf(
			"Failed to parse --page/-p value %q, error: %v", rawOpts.Display.PageSize, err)
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
