package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/out"
)

type Bool struct {
	reader       io.Reader
	printer      *out.Printer
	message      string
	defaultValue bool
}

func NewBool(message string, defaultValue bool) *Bool {
	return &Bool{
		reader:       os.Stdin,
		printer:      out.Default,
		message:      message,
		defaultValue: defaultValue,
	}
}

func (b *Bool) Prompt() (bool, error) {
	b.printer.Msgf("%s %s: ", b.message, b.defaultValueDisplay())

	scanner := bufio.NewScanner(b.reader)
	scanner.Scan()
	err := scanner.Err()
	userInput := scanner.Text()
	b.printer.Msg("\n")

	if err != nil {
		return false, err
	}

	parsedUserInput, err := b.validateUserInput(userInput)
	if err != nil {
		return false, err
	}

	return parsedUserInput, nil
}

func (b *Bool) defaultValueDisplay() string {
	if b.defaultValue {
		return "[Y/n]"
	}
	return "[y/N]"
}

func (b *Bool) validateUserInput(userInput string) (bool, error) {
	switch strings.TrimSpace(strings.ToLower(userInput)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return b.defaultValue, nil
	default:
		return false, fmt.Errorf("invalid input, please enter 'y' or 'n'")
	}
}
