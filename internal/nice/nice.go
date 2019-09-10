// package nice provides nice printing for the CLI
package nice

import (
	"fmt"

	ct "github.com/daviddengcn/go-colortext"
)

// PrintKyma prints the Kyma word with its identity color
func PrintKyma() {
	ct.ChangeColor(ct.Cyan, false, ct.None, false)
	fmt.Print("Kyma")
	ct.ResetColor()
}

func PrintImportant(s string) {
	ct.ChangeColor(ct.Yellow, true, ct.None, false)
	fmt.Println(s)
	ct.ResetColor()
}

func PrintImportantf(format string, a ...interface{}) {
	PrintImportant(fmt.Sprintf(format, a...))
}
