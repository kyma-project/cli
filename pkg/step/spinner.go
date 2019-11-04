package step

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/root"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func newStepWithSpinner(msg string) Step {
	s := spinner.New(
		[]string{"/", "-", "\\", "|"},
		time.Millisecond*200,
		spinner.WithColor("reset"),
		spinner.WithSuffix(" "+msg),
	)
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
	s.spinner.Suffix = fmt.Sprintf(" %s: %s", s.msg, msg)
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
		gliph = color.GreenString(successGlyph)
	} else {
		gliph = color.RedString(failureGlyph)
	}
	s.spinner.FinalMSG = fmt.Sprintf("%s%s\n", gliph, s.msg)
	s.spinner.Stop()
}

func (s *stepWithSpinner) LogInfo(msg string) {
	s.logTo(os.Stdout, infoGlyph+msg)
}

func (s *stepWithSpinner) LogInfof(format string, args ...interface{}) {
	s.logTof(os.Stdout, infoGlyph+format, args...)
}

func (s *stepWithSpinner) LogError(msg string) {
	s.logTof(os.Stderr, color.YellowString(warningGlyph)+msg)
}

func (s *stepWithSpinner) LogErrorf(format string, args ...interface{}) {
	s.logTof(os.Stderr, color.YellowString(warningGlyph)+format, args...)
}

func (s *stepWithSpinner) logTof(to io.Writer, format string, args ...interface{}) {
	isActive := s.spinner.Active()
	s.spinner.Stop()
	fmt.Fprintf(to, format+"\n", args...)
	if isActive {
		s.spinner.Start()
	}
}

func (s *stepWithSpinner) logTo(to io.Writer, format string) {
	isActive := s.spinner.Active()
	s.spinner.Stop()
	fmt.Fprint(to, format+"\n")
	if isActive {
		s.spinner.Start()
	}
}

func (s *stepWithSpinner) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	isActive := s.spinner.Active()
	s.spinner.Stop()
	fmt.Printf("%s%s", questionGlyph, msg)
	answer, err := reader.ReadString('\n')
	if isActive {
		s.spinner.Start()
	}
	return strings.TrimSpace(answer), err
}

func (s *stepWithSpinner) PromptYesNo(msg string) bool {
	isActive := s.spinner.Active()
	s.spinner.Stop()
	fmt.Printf("%s%s", questionGlyph, msg)
	answer := root.PromptUser()
	if isActive {
		s.spinner.Start()
	}
	return answer
}
