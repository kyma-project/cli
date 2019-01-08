package step

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

func NewStepWithSpinner(msg string) Step {
	s := spinner.New([]string{"/", "-", "\\", "|"}, time.Millisecond * 200)
	s.Prefix = msg+" "
	return  &stepWithSpinner{s, msg}
}

type stepWithSpinner struct {
	spinner *spinner.Spinner
	msg string
}

func (s *stepWithSpinner) Start() {
	s.spinner.Start()
}

func (s *stepWithSpinner) Status(msg string) {
	s.spinner.Suffix = fmt.Sprintf(" : %s", msg)
}

func (s *stepWithSpinner) Success() {
	s.Stop(true)
}

func (s *stepWithSpinner) Successf(format string, args ...interface{}) {
	s.Stopf(true, format, args...)
}

func (s *stepWithSpinner) Failure() {
	s.Stop(false)
}

func (s *stepWithSpinner) Failuref(format string, args ...interface{}) {
	s.Stopf(false, format, args...)
}

func (s *stepWithSpinner) Stopf(success bool, format string, args ...interface{}) {
	s.msg = fmt.Sprintf(format, args...)
	s.Stop(success)
}

func (s *stepWithSpinner) Stop(success bool) {
	var gliph string
	if success {
		gliph = successGliph
	} else {
		gliph = failureGliph
	}
	s.spinner.FinalMSG = fmt.Sprintf("%s %s\n", s.msg, gliph)
	s.spinner.Stop()
}
