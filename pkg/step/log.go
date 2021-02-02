package step

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/root"
)

func newLogStep(msg string) Step {
	return &logStep{msg}
}

type logStep struct {
	msg string
}

func (s *logStep) Start() {
	log.Println(s.msg)
}

func (s *logStep) Status(msg string) {
	log.Printf("%s: %s\n", s.msg, msg)
}

func (s *logStep) Success() {
	s.Stop(true)
}

func (s *logStep) Successf(format string, args ...interface{}) {
	s.Stopf(true, format, args...)
}

func (s *logStep) Failure() {
	s.Stop(false)
}

func (s *logStep) Failuref(format string, args ...interface{}) {
	s.Stopf(false, format, args...)
}

func (s *logStep) Stopf(success bool, format string, args ...interface{}) {
	s.msg = fmt.Sprintf(format, args...)
	s.Stop(success)
}

func (s *logStep) Stop(success bool) {
	log.Println(s.msg)
}

func (s *logStep) LogInfo(msg string) {
	log.Println(msg)
}

func (s *logStep) LogInfof(format string, args ...interface{}) {
	s.LogInfo(fmt.Sprintf(format, args...))
}

func (s *logStep) LogError(msg string) {
	log.Println(msg)
}

func (s *logStep) LogErrorf(format string, args ...interface{}) {
	s.LogError(fmt.Sprintf(format, args...))
}

func (s *logStep) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	log.Print(msg)
	answer, err := reader.ReadString('\n')
	return strings.TrimSpace(answer), err
}

func (s *logStep) PromptYesNo(msg string) bool {
	log.Print(msg)
	answer := root.PromptUser()
	return answer
}

func (s *logStep) String() string {
	return s.msg
}
