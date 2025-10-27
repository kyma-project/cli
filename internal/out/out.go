// Package out standardizes output handling across the CLI.
// It provides a few levels of output represented by different functions:
// - General output messages (Msg, Msgf, Msgln, Msgfln) write to stdout by default and can be disabled
// - Priority messages (Prio, Priof, Prioln, Priofln) write to stdout by default and cannot be disabled
// - Verbose messages (Verbose, Verbosef, Verboseln, Verbosefln) write to stdout when enabled
// - Debug messages (Dev, Devf, Devln, Devfln) write to stdout when enabled
// - Error messages (Err, Errf, Errln, Errfln) write to stderr by default
// Printer can be configured with different writers to redirect output as needed to files, buffers, etc.
package out

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Default is the default global Printer instance used by package-level functions
var Default = New()

func SetForCmd(cmd *cobra.Command) { Default.SetForCmd(cmd) }
func DisableMsg()                  { Default.DisableMsg() }
func EnableVerbose()               { Default.EnableVerbose() }
func EnableDebug()                 { Default.EnableDebug() }

func Msg(msg string)                         { Default.Msg(msg) }
func Msgln(msg string)                       { Default.Msgln(msg) }
func Msgf(format string, a ...interface{})   { Default.Msgf(format, a...) }
func Msgfln(format string, a ...interface{}) { Default.Msgfln(format, a...) }

func Prio(msg string)                       { Default.Prio(msg) }
func Prioln(msg string)                     { Default.Prioln(msg) }
func Priof(format string, a ...interface{}) { Default.Priof(format, a...) }

func Verbose(msg string)                         { Default.Verbose(msg) }
func Verboseln(msg string)                       { Default.Verboseln(msg) }
func Verbosef(format string, a ...interface{})   { Default.Verbosef(format, a...) }
func Verbosefln(format string, a ...interface{}) { Default.Verbosefln(format, a...) }

func Debug(msg string)                         { Default.Debug(msg) }
func Debugln(msg string)                       { Default.Debugln(msg) }
func Debugf(format string, a ...interface{})   { Default.Debugf(format, a...) }
func Debugfln(format string, a ...interface{}) { Default.Debugfln(format, a...) }

func Err(msg string)                         { Default.Err(msg) }
func Errln(msg string)                       { Default.Errln(msg) }
func Errf(format string, a ...interface{})   { Default.Errf(format, a...) }
func Errfln(format string, a ...interface{}) { Default.Errfln(format, a...) }

type Printer struct {
	outWriter        io.Writer
	prioOutWriter    io.Writer
	verboseOutWriter io.Writer
	errWriter        io.Writer
	debugWriter      io.Writer
}

// New creates a new Printer with default settings
// normal output to stdout, error output to stderr, debug and verbose outputs disabled
func New() *Printer {
	return &Printer{
		outWriter:        os.Stdout,
		errWriter:        os.Stderr,
		prioOutWriter:    os.Stdout,
		verboseOutWriter: io.Discard,
		debugWriter:      io.Discard,
	}
}

// NewToWriter creates a new Printer that uses the same writer for all output types
// can be used to write all output to a file or buffer
func NewToWriter(writer io.Writer) *Printer {
	return &Printer{
		outWriter:        writer,
		errWriter:        writer,
		prioOutWriter:    writer,
		verboseOutWriter: io.Discard,
		debugWriter:      io.Discard,
	}
}

// SetForCmd sets the output and error writers based on the cobra command context
func (p *Printer) SetForCmd(cmd *cobra.Command) {
	p.prioOutWriter = cmd.OutOrStdout()
	p.outWriter = cmd.OutOrStdout()
	p.errWriter = cmd.ErrOrStderr()
}

// DisableMsg disables normal output messages
// use this to suppress normal output when needed (e.g. in scripts)
// Note: Prio, Dev and Err messages are not affected
func (p *Printer) DisableMsg() {
	p.outWriter = io.Writer(io.Discard)
}

// EnableDebug enables debug output messages
func (p *Printer) EnableDebug() {
	p.debugWriter = p.prioOutWriter
}

// EnableVerbose enables verbose output messages
func (p *Printer) EnableVerbose() {
	p.verboseOutWriter = p.outWriter
}

// Msg prints message to the output writer
// alternative for the fmt.Print functions
func (p *Printer) Msg(msg string) {
	write(p.outWriter, "%s", msg)
}

// Msgln prints a message to the output writer
// alternative for the fmt.Println function
func (p *Printer) Msgln(msg string) {
	writeln(p.outWriter, "%s", msg)
}

// Msgf prints a formatted message to the output writer
// alternative for the fmt.Printf functions
func (p *Printer) Msgf(format string, a ...interface{}) {
	write(p.outWriter, format, a...)
}

// Msgfln prints a formatted message to the output writer with newline
// combination of fmt.Printf and fmt.Println
func (p *Printer) Msgfln(format string, a ...interface{}) {
	writeln(p.outWriter, format, a...)
}

// Prio prints a message to the priority output writer
// alternative for the fmt.Print functions but for high priority messages
func (p *Printer) Prio(msg string) {
	write(p.prioOutWriter, "%s", msg)
}

// Prioln prints a message to the priority output writer
// alternative for the fmt.Println function but for high priority messages
func (p *Printer) Prioln(msg string) {
	writeln(p.prioOutWriter, "%s", msg)
}

// Priof prints a formatted message to the priority output writer
// alternative for the fmt.Printf functions but for high priority messages
func (p *Printer) Priof(format string, a ...interface{}) {
	write(p.prioOutWriter, format, a...)
}

// Priofln prints a formatted message to the priority output writer with newline
// combination of fmt.Printf and fmt.Println but for high priority messages
func (p *Printer) Priofln(format string, a ...interface{}) {
	writeln(p.prioOutWriter, format, a...)
}

// Verbose prints a verbose message to the verbose output writer
// alternative for the fmt.Print functions but for verbose messages
func (p *Printer) Verbose(msg string) {
	write(p.verboseOutWriter, "%s", msg)
}

// Verboseln prints a verbose message to the verbose output writer
// alternative for the fmt.Println function but for verbose messages
func (p *Printer) Verboseln(msg string) {
	writeln(p.verboseOutWriter, "%s", msg)
}

// Verbosef prints a formatted verbose message to the verbose output writer
// alternative for the fmt.Printf functions but for verbose messages
func (p *Printer) Verbosef(format string, a ...interface{}) {
	write(p.verboseOutWriter, format, a...)
}

// Verbosefln prints a formatted verbose message to the verbose output writer with newline
// combination of fmt.Printf and fmt.Println but for verbose messages
func (p *Printer) Verbosefln(format string, a ...interface{}) {
	writeln(p.verboseOutWriter, format, a...)
}

// Debug prints a debug message to the debug output writer
// alternative for the fmt.Print functions but for debug messages
func (p *Printer) Debug(msg string) {
	write(p.debugWriter, "%s", msg)
}

// Debugln prints a debug message to the debug output writer
// alternative for the fmt.Println function but for debug messages
func (p *Printer) Debugln(msg string) {
	writeln(p.debugWriter, "%s", msg)
}

// Debugf prints a formatted debug message to the debug output writer
// alternative for the fmt.Printf functions but for debug messages
func (p *Printer) Debugf(format string, a ...interface{}) {
	write(p.debugWriter, format, a...)
}

// Debugfln prints a formatted debug message to the debug output writer with newline
// combination of fmt.Printf and fmt.Println but for debug messages
func (p *Printer) Debugfln(format string, a ...interface{}) {
	writeln(p.debugWriter, format, a...)
}

// Err prints a error message to the error output writer
// alternative for the fmt.Print functions but for error messages
func (p *Printer) Err(msg string) {
	write(p.errWriter, "%s", msg)
}

// Errln prints a error message to the error output writer
// alternative for the fmt.Println function but for error messages
func (p *Printer) Errln(msg string) {
	writeln(p.errWriter, "%s", msg)
}

// Errf prints a formatted error message to the error output writer
// alternative for the fmt.Printf functions but for error messages
func (p *Printer) Errf(format string, a ...interface{}) {
	write(p.errWriter, format, a...)
}

// Errfln prints a formatted error message to the error output writer with newline
// combination of fmt.Printf and fmt.Println but for error messages
func (p *Printer) Errfln(format string, a ...interface{}) {
	writeln(p.errWriter, format, a...)
}

// Write implements the io.Writer interface for Printer
// it allows using Printer as an io.Writer
func (p *Printer) Write(b []byte) (n int, err error) {
	return p.MsgWriter().Write(b)
}

// MsgWriter returns the output writer
// for passing it to functions that require an io.Writer
// in this way the output can be muted using DisableMsg()
func (p *Printer) MsgWriter() io.Writer {
	return p.outWriter
}

// ErrWriter returns the error output writer
// for passing it to functions that require an io.Writer
func (p *Printer) ErrWriter() io.Writer {
	return p.errWriter
}

func write(writer io.Writer, format string, a ...interface{}) {
	fmt.Fprintf(writer, format, a...)
}

func writeln(writer io.Writer, format string, a ...interface{}) {
	fmt.Fprintf(writer, fmt.Sprintf("%s\n", format), a...)
}
