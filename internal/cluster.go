package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/clebs/ldng"
)

const (
	sleep = 5 * time.Second
)

// RunCmd executes a command with given arguments
func RunCmd(c string, args []string) (string, error) {
	cmd := exec.Command(c, args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing command '%s %s' with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

//NewSpinner starts and returns a new terminal spinner using given text as progress and success text
func NewSpinner(startText string, stopText string) chan struct{} {
	s := ldng.NewSpin(ldng.SpinPrefix(fmt.Sprintf("%s", startText)), ldng.SpinPeriod(100*time.Millisecond), ldng.SpinSuccess(fmt.Sprintf("%s ✅\n", stopText)))
	return s.Start()
}

//StopSpinner marks a terminal spinner as finished with success
func StopSpinner(spinner chan struct{}) {
	spinner <- struct{}{} // stop the spinner
	<-spinner             // wait for the spinner to finish cleanup
	spinner = nil
}
