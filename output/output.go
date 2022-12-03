package output

import (
	"bufio"
	"fmt"
	"io"

	"github.com/jcuga/hax/commands"
	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Output(writer io.Writer, reader *input.FixedLengthBufferedReader, ioInfo options.IOInfo,
	opts options.Options, cmd options.Command, cmdArgs []string) error {
	// use buffered writer for better performance.
	// ex: displaying line by line to stdout or a file when displaying hex is slow.
	// buffering the writes significantly speeds up the hex display output.
	w := bufio.NewWriter(writer)

	defer func() {
		if !ioInfo.StdoutIsPipe {
			// add newline to end of terminal output
			fmt.Fprintf(w, "\n")
		}
		// always flush output writer!
		// NOTE: used to have os.Exit within various func calls below which DOES NOT call
		// defered statements. Now returnign errors back to main so this defer fires--TIL.
		w.Flush()
	}()

	if cmd != options.NoCommand {
		switch cmd {
		case options.CountBytes:
			return commands.CountBytes(w, reader, ioInfo, opts, cmdArgs)
		case options.Strings:
			return commands.Strings(w, reader, ioInfo, opts, cmdArgs)
		case options.StringsUtf8:
			return commands.StringsUtf8(w, reader, ioInfo, opts, cmdArgs)
		case options.Search:
			return commands.Search(w, reader, ioInfo, opts, cmdArgs)
		default:
			return fmt.Errorf("Unhandled command: %q", options.CommandToString(cmd))
		}
	}

	switch opts.OutputMode {
	case options.Base64:
		return outputBase64(w, reader, ioInfo, opts)
	case options.Display:
		return displayHex(w, reader, ioInfo, opts)
	case options.Hex:
		return outputHex(w, reader, ioInfo, opts)
	case options.HexString:
		return outputHexStringOrList(w, reader, ioInfo, opts)
	case options.HexList:
		return outputHexStringOrList(w, reader, ioInfo, opts)
	case options.HexAscii:
		return outputHexAscii(w, reader, ioInfo, opts)
	case options.Raw:
		return outputRaw(w, reader, ioInfo, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
}
