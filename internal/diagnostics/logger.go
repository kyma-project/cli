package diagnostics

import (
	"fmt"
	"io"
)

type VerboseLogger struct {
	writer  io.Writer
	verbose bool
}

func NewVerboseLogger(writer io.Writer, verbose bool) VerboseLogger {
	return VerboseLogger{
		writer:  writer,
		verbose: verbose,
	}
}

func (vl *VerboseLogger) WriteVerboseError(err error, message string) {
	if !vl.verbose || err == nil {
		return
	}

	fmt.Fprintf(vl.writer, "%s: %s\n", message, err.Error())
}
