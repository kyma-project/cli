package deploy

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/step"
)

//StepFactory abstracts an entity which creates
type StepFactory interface {
	AddStep(stepName string) step.Step
}

type uiStep struct {
	logFunc func(string, ...interface{})
	msg     string
	success bool
	running bool
}

//NewUIStep creates a new UI step
func NewUIStep(msg string, logFunc func(format string, args ...interface{})) step.Step {
	return &uiStep{
		msg:     msg,
		logFunc: logFunc,
	}
}

func (s *uiStep) Start() {
	s.running = true
	s.logFunc("%s", s.msg)
}

func (s *uiStep) Status(msg string) {
	s.logFunc("%s - running:%s, success:%s", msg, s.running, s.success)
}

func (s *uiStep) Success() {
	s.success = true
	s.logFunc("%s: Ok", s.msg)
}

func (s *uiStep) Successf(format string, args ...interface{}) {
	s.success = true
	s.logFunc(format, args...)
}

func (s *uiStep) Failure() {
	s.success = false
	s.logFunc("%s: Failed", s.msg)
}

func (s *uiStep) Failuref(format string, args ...interface{}) {
	s.success = false
	s.logFunc(format, args...)
}

func (s *uiStep) Stop(success bool) {
	s.success = success
	s.running = false
	s.logFunc("%s stopped", s.msg)
}

func (s *uiStep) Stopf(success bool, format string, args ...interface{}) {
	s.success = success
	s.running = false
	s.logFunc(format, args...)
}

func (s *uiStep) LogInfo(msg string) {
	s.logFunc(msg)
}

func (s *uiStep) LogInfof(format string, args ...interface{}) {
	s.logFunc(format, args...)
}

func (s *uiStep) LogError(msg string) {
	s.logFunc(msg)
}

func (s *uiStep) LogErrorf(format string, args ...interface{}) {
	s.logFunc(format, args...)
}

func (s *uiStep) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", msg)
	answer, err := reader.ReadString('\n')
	return strings.TrimSpace(answer), err
}

func (s *uiStep) PromptYesNo(msg string) bool {
	fmt.Printf("%s ", msg)
	answer := root.PromptUser()
	return answer
}
