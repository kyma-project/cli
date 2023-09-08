// Package nice provides nice printing for the CLI
package nice

import (
	"fmt"

	ct "github.com/daviddengcn/go-colortext"
)

// Nice contains the option to determine the interactivity of the printing.
type Nice struct {
	NonInteractive bool
}

// NewNice returns a new Nice instance
func NewNice(NonInteractive bool) *Nice {
	return &Nice{NonInteractive}
}

// PrintKyma prints the Kyma word with its identity color
func (n *Nice) PrintKyma() {
	if n.NonInteractive {
		fmt.Print("Kyma")
	} else {
		ct.ChangeColor(ct.Cyan, false, ct.None, false)
		fmt.Print("Kyma")
		ct.ResetColor()
	}
}

// PrintImportant prints the line with yellow color
func (n *Nice) PrintImportant(s string) {
	if n.NonInteractive {
		fmt.Println(s)
	} else {
		ct.ChangeColor(ct.Yellow, true, ct.None, false)
		fmt.Println(s)
		ct.ResetColor()
	}
}

// PrintImportantf prints the line with yellow color
func (n *Nice) PrintImportantf(format string, a ...interface{}) {
	n.PrintImportant(fmt.Sprintf(format, a...))
}
