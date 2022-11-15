package output

import (
	"bufio"
	"fmt"
	"io"

	"github.com/jcuga/hax/commands"
	"github.com/jcuga/hax/input"
	"github.com/jcuga/hax/options"
)

func Output(writer io.Writer, reader *input.FixedLengthBufferedReader, isPipe, isStdin bool,
	opts options.Options, cmd options.Command, cmdArgs []string) error {
	// use buffered writer for better performance.
	// ex: displaying line by line to stdout or a file when displaying hex is slow.
	// buffering the writes significantly speeds up the hex display output.
	w := bufio.NewWriter(writer)
	defer w.Flush()

	if cmd != options.NoCommand {
		switch cmd {
		case options.Strings:
			commands.Strings(w, reader, isPipe, isStdin, opts, cmdArgs)
			return nil
		case options.CountBytes:
			commands.CountBytes(w, reader, isPipe, isStdin, opts, cmdArgs)
			return nil
		default:
			return fmt.Errorf("Unhandled command: %q", options.CommandToString(cmd))
		}
	}

	switch opts.OutputMode {
	case options.Base64:
		outputBase64(w, reader, isPipe, isStdin, opts)
	case options.Display:
		displayHex(w, reader, isPipe, isStdin, opts)
	case options.Hex:
		outputHex(w, reader, isPipe, isStdin, opts)
	case options.HexString:
		outputHexStringOrList(w, reader, isPipe, isStdin, opts)
	case options.HexList:
		outputHexStringOrList(w, reader, isPipe, isStdin, opts)
	case options.HexAscii:
		outputHexAscii(w, reader, isPipe, isStdin, opts)
	case options.Raw:
		outputRaw(w, reader, isPipe, isStdin, opts)
	default:
		return fmt.Errorf("Unsupported or not implemented output mode: %v", opts.OutputMode)
	}
	return nil
}
