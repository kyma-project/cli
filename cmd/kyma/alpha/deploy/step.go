package deploy

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/step"
)

type deploymentStep struct {
	logFunc func(string, ...interface{})
	msg     string
	success bool
	running bool
}

func newDeploymentStep(msg string, logFunc func(format string, args ...interface{})) step.Step {
	return &deploymentStep{
		msg:     msg,
		logFunc: logFunc,
	}
}

func (s *deploymentStep) Start() {
	s.running = true
	s.logFunc("%s", s.msg)
}

func (s *deploymentStep) Status(msg string) {
	s.logFunc("%s - running:%s, success:%s", msg, s.running, s.success)
}

func (s *deploymentStep) Success() {
	s.success = true
	s.logFunc("%s: Ok", s.msg)
}

func (s *deploymentStep) Successf(format string, args ...interface{}) {
	s.success = true
	s.logFunc(format, args...)
}

func (s *deploymentStep) Failure() {
	s.success = false
	s.logFunc("%s: Failed", s.msg)
}

func (s *deploymentStep) Failuref(format string, args ...interface{}) {
	s.success = false
	s.logFunc(format, args...)
}

func (s *deploymentStep) Stop(success bool) {
	s.success = success
	s.running = false
	s.logFunc("%s stopped", s.msg)
}

func (s *deploymentStep) Stopf(success bool, format string, args ...interface{}) {
	s.success = success
	s.running = false
	s.logFunc(format, args...)
}

func (s *deploymentStep) LogInfo(msg string) {
	s.logFunc(msg)
}

func (s *deploymentStep) LogInfof(format string, args ...interface{}) {
	s.logFunc(format, args...)
}

func (s *deploymentStep) LogError(msg string) {
	s.logFunc(msg)
}

func (s *deploymentStep) LogErrorf(format string, args ...interface{}) {
	s.logFunc(format, args...)
}

func (s *deploymentStep) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", msg)
	answer, err := reader.ReadString('\n')
	return strings.TrimSpace(answer), err
}

func (s *deploymentStep) PromptYesNo(msg string) bool {
	fmt.Printf("%s ", msg)
	answer := root.PromptUser()
	return answer
}
