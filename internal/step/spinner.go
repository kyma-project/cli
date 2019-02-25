package step

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

func NewStepWithSpinner(msg string) Step {
	s := spinner.New(
		[]string{"/", "-", "\\", "|"},
		time.Millisecond*200,
		spinner.WithColor("reset"),
	)
	s.Prefix = msg + " "
	return &stepWithSpinner{s, msg}
}

type stepWithSpinner struct {
	spinner *spinner.Spinner
	msg     string
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
	s.spinner.FinalMSG = fmt.Sprintf("%s %s\n", gliph, s.msg)
	s.spinner.Stop()
}

func (s *stepWithSpinner) LogInfo(msg string) {
	s.logTof(os.Stdout, infoGliph+"  "+msg)
}

func (s *stepWithSpinner) LogInfof(format string, args ...interface{}) {
	s.logTof(os.Stdout, infoGliph+"  "+format, args...)
}

func (s *stepWithSpinner) LogError(msg string) {
	s.logTof(os.Stderr, warningGliph+"  "+msg)
}

func (s *stepWithSpinner) LogErrorf(format string, args ...interface{}) {
	s.logTof(os.Stderr, warningGliph+"  "+format, args)
}

func (s *stepWithSpinner) logTof(to io.Writer, format string, args ...interface{}) {
	isActive := s.spinner.Active()
	s.spinner.Stop()
	_, _ = fmt.Fprintf(to, format+"\n", args...)
	if isActive {
		s.spinner.Start()
	}
}

func (s *stepWithSpinner) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	isActive := s.spinner.Active()
	s.spinner.Stop()
	fmt.Printf("%s %s", questionGliph, msg)
	answer, err := reader.ReadString('\n')
	if isActive {
		s.spinner.Start()
	}
	return strings.TrimSpace(answer), err
}
