package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
	"github.com/jcuga/hax/output"
)

func main() {
	// Input is either: file, --str arg, or stdin
	var inFilename string
	flag.StringVar(&inFilename, "file", "", "Filename to read from. Parsed as raw by default (change with --input).")
	flag.StringVar(&inFilename, "f", "", "")
	var inputStr string
	flag.StringVar(&inputStr, "str", "", "String input instead of file/stdin. Parsed as hex by default.")
	flag.StringVar(&inputStr, "s", "", "")

	// Input/output can be raw binary, hex string, base64 string, binary string,
	// or copy+paste of hax output.
	var inMode string
	flag.StringVar(&inMode, "input", "", "Input mode. See I/O Modes section for options.")
	flag.StringVar(&inMode, "i", "", "")
	var outMode string
	flag.StringVar(&outMode, "output", "", "Output mode. (defaults to hexeditor formatted output).")
	flag.StringVar(&outMode, "o", "", "")

	// Optional input limit/offset. This can be in decimal or hex.
	var offset string
	flag.StringVar(&offset, "offset", "", "Input offset in bytes (default 0).")
	flag.StringVar(&offset, "n", "", "")
	var limit string
	flag.StringVar(&limit, "limit", "", "Input limit in bytes (default no limit).")
	flag.StringVar(&limit, "l", "", "")

	// Customize display mode output:
	var colWidth string
	// TODO: don't have a default, then default based on output mode if not specified.
	// TODO: update -h/usage output to reflect this change.
	flag.StringVar(&colWidth, "width", "16", "Column Width: Num bytes per row (default: 16).")
	flag.StringVar(&colWidth, "w", "16", "")
	var pageSize string
	flag.StringVar(&pageSize, "page", "4", "Display page breaks every N (default 4, 0=never).")
	flag.StringVar(&pageSize, "p", "4", "")
	var alwaysPretty bool
	flag.BoolVar(&alwaysPretty, "pretty", false, "Always pretty-print/style output.")
	flag.BoolVar(&alwaysPretty, "y", false, "")
	var quiet bool
	flag.BoolVar(&quiet, "no-ascii", false, "Skip outputting ascii below each row of bytes.")
	flag.BoolVar(&quiet, "q", false, "")

	flag.Usage = func() {
		w := flag.CommandLine.Output() // may be os.Stderr - but not necessarily
		// NOTE: custom stuff before
		fmt.Fprintf(w, "HAX - binary pocket knife.\n\n")
		fmt.Fprintf(w, "Usage:\thax [input options] [output options] [cmd]\n")

		fmt.Fprintf(w, "\nInput Options:\n")
		f := flag.Lookup("file")
		fmt.Fprintf(w, "\t-f, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("str")
		fmt.Fprintf(w, "\t-s, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("input")
		fmt.Fprintf(w, "\t-i, --%s\t%s\n", f.Name, f.Usage)
		fmt.Fprintf(w, "\t\t\tIf stdin or --file input, defaults to raw, if --str defaults to hex.\n")

		f = flag.Lookup("offset")
		fmt.Fprintf(w, "\t-n, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("limit")
		fmt.Fprintf(w, "\t-l, --%s\t%s\n", f.Name, f.Usage)

		fmt.Fprintf(w, "\nOutput Options:\n")
		f = flag.Lookup("output")
		fmt.Fprintf(w, "\t-o, --%s\t%s\n", f.Name, f.Usage)

		fmt.Fprintf(w, "\nOptions for when output mode is display:\n")
		f = flag.Lookup("width")
		fmt.Fprintf(w, "\t-w, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("page")
		fmt.Fprintf(w, "\t-p, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("no-ascii")
		fmt.Fprintf(w, "\t-q, --%s\t%s\n", f.Name, f.Usage)
		f = flag.Lookup("pretty")
		fmt.Fprintf(w, "\t-y, --%s\t%s\n", f.Name, f.Usage)

		fmt.Fprintf(w, "\nI/O Modes:\n")
		fmt.Fprintf(w, "  * r, raw\tRaw bytes.\n")
		fmt.Fprintf(w, "  * h, hex\tHex string.\n")
		fmt.Fprintf(w, "  * b, base64\tBase64 string.\n")

		fmt.Fprintf(w, "\nNote:\n")
		fmt.Fprintf(w, "  * If no --file or --str set, will get input from stdin.\n")
		fmt.Fprintf(w, "  * For any numeric args (ex: limit, offset, etc), values starting with:\n")
		fmt.Fprintf(w, "      '0', '0x', '\\x', or, 'x' are parsed as hex instead of decimal.\n")
		fmt.Fprintf(w, "    Same goes if value contains A-F or a-f.\n")

		fmt.Fprintf(w, "\nTODO: optional commands like conv to num, str, unicode, binary, math, etc.\n")

		fmt.Fprintf(w, "\nExamples:\n\nTodo use -e, --examples to see examples\n")
	}

	flag.Parse()

	// Don't allow any positional args--currently everything is a flag.
	// TODO: update this to only enforce when given opts don't allow positional
	// if/when opts with positional args are added.
	if flag.NArg() > 0 {
		fmt.Printf("Unhandled positional args (%d)\n", flag.NArg())
		flag.Usage()
		os.Exit(1)
	}

	opts, err := options.New(inFilename, inputStr, inMode, outMode,
		offset, limit, colWidth, pageSize, alwaysPretty, quiet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	inReader, inCloser, err := input.GetInput(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if inCloser != nil {
		defer inCloser.Close()
	}

	// fmt.Printf("Parsed Options: %v\n", opts) // TODO: remove me

	isPipe := false
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		isPipe = true
	}

	// TODO: do this only if there's no other cmd/func (ex: interpret as numeric, insert, replace, etc)
	if err := output.Output(inReader, isPipe, opts); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// TODO: implement various command/utility funcs (parse numeric, str, unicode, math)
	// TODO: implement edit/insert/replace contents
	// TODO: binary level stuff... display, maths...
}
