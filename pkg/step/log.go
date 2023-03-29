package step

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/root"
	"go.uber.org/zap"
)

func newLogStep(msg string) Step {
	logger, err := zap.NewDevelopment()
	if err != nil {
		if err != nil {
			log.Fatalf("Can't initialize zap logger: %v", err)
		}
	}
	return &logStep{
		msg:    msg,
		logger: logger,
	}
}

type logStep struct {
	msg    string
	logger *zap.Logger
}

func (s *logStep) Start() {
	s.logger.Info(s.msg)
}

func (s *logStep) Status(msg string) {
	s.logger.Info(fmt.Sprintf("%s: %s\n", s.msg, msg))
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

func (s *logStep) Stop(_ bool) {
	s.logger.Info(s.msg)
}

func (s *logStep) LogInfo(msg string) {
	s.logger.Info(msg)
}

func (s *logStep) LogInfof(format string, args ...interface{}) {
	s.LogInfo(fmt.Sprintf(format, args...))
}

func (s *logStep) LogError(msg string) {
	s.logger.Error(msg)
}

func (s *logStep) LogErrorf(format string, args ...interface{}) {
	s.LogError(fmt.Sprintf(format, args...))
}

func (s *logStep) LogWarn(msg string) {
	s.logger.Warn(msg)
}

func (s *logStep) LogWarnf(format string, args ...interface{}) {
	s.LogWarn(fmt.Sprintf(format, args...))
}

func (s *logStep) Prompt(msg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	s.logger.Info(msg)
	answer, err := reader.ReadString('\n')
	return strings.TrimSpace(answer), err
}

func (s *logStep) PromptYesNo(msg string) bool {
	s.logger.Info(msg)
	answer := root.PromptUser()
	return answer
}

func (s *logStep) String() string {
	return s.msg
}
