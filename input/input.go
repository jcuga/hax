package input

import (
	"fmt"
	"io"

	"github.com/jcuga/hax/options"
)

// TODO: get file, stdin or str input
// TODO: for file and stdin, stream it!

func GetReader(opts options.Options) (io.Reader, error) {
	if len(opts.InputData) > 0 {

	} else if len(opts.Filename) > 0 {

	} else {
		// TODO: stdin
	}

	return nil, fmt.Errorf("not implemented")
}
