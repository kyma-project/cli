package step

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/root"
)

func newSimpleStep(msg string) Step {
	return &simpleStep{msg}
}

type simpleStep struct {
	msg string
}

func (s *simpleStep) Start() {
	fmt.Println(s.msg)
}

func (s *simpleStep) Status(msg string) {
	fmt.Printf("%s: %s\n", s.msg, msg)
}

func (s *simpleStep) Success() {
	s.Stop(true)
}

func (s *simpleStep) Successf(format string, args ...interface{}) {
	s.Stopf(true, format, args...)
}

func (s *simpleStep) Failure() {
	s.Stop(false)
}

func (s *simpleStep) Failuref(format string, args ...interface{}) {
	s.Stopf(false, format, args...)
}

func (s *simpleStep) Stopf(success bool, format string, args ...interface{}) {
	s.msg = fmt.Sprintf(format, args...)
	s.Stop(success)
}

func (s *simpleStep) Stop(success bool) {
	var glyph string
	if success {
		glyph = successGlyph
	} else {
		glyph = failureGlyph
	}
	fmt.Printf("%s%s\n", glyph, s.msg)
}

func (s *simpleStep) LogInfo(msg string) {
	fmt.Printf("%s%s\n", infoGlyph, msg)
}

func (s *simpleStep) LogInfof(format string, args ...interface{}) {
	s.LogInfo(fmt.Sprintf(format, args...))
}

func (s *simpleStep) LogError(msg string) {
	fmt.Fprintf(os.Stderr, "%s%s\n", warningGlyph, msg)
}

func (s *simpleStep) LogErrorf(format string, args ...interface{}) {
	s.LogError(fmt.Sprintf(format, args...))
}

func (s *simpleStep) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s%s", questionGlyph, msg)
	answer, err := reader.ReadString('\n')
	return strings.TrimSpace(answer), err
}

func (s *simpleStep) PromptYesNo(msg string) bool {
	fmt.Printf("%s%s", questionGlyph, msg)
	answer := root.PromptUser()
	return answer
}
