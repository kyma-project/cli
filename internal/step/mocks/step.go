package mocks

import "fmt"

// Mock of a CLI step.
// All logged messages and status are stored and can be retreived later for validation.
type Step struct {
	status, infos, errs []string
	success, stopped    bool
}

func (s *Step) Start() {

}

func (s *Step) Status(msg string) {
	s.status = append(s.status, msg)
}

func (s *Step) Statuses() []string {
	return s.status
}

func (s *Step) Success() {
	s.success = true
}

func (s *Step) Successf(format string, args ...interface{}) {
	s.status = append(s.status, fmt.Sprintf(format, args...))
	s.Success()
}

func (s *Step) IsSuccessful() bool {
	return s.success
}

func (s *Step) Failure() {
	s.success = false
}

func (s *Step) Failuref(format string, args ...interface{}) {
	s.status = append(s.status, fmt.Sprintf(format, args...))
	s.Failure()
}

func (s *Step) Stop(success bool) {
	s.stopped = true
	s.success = success
}

func (s *Step) Stopf(success bool, format string, args ...interface{}) {
	s.status = append(s.status, fmt.Sprintf(format, args...))
	s.Stop(success)
}

func (s *Step) IsStopped() bool {
	return s.stopped
}

func (s *Step) LogInfo(msg string) {
	s.infos = append(s.infos, msg)
}

func (s *Step) LogInfof(format string, args ...interface{}) {
	s.LogInfo(fmt.Sprintf(format, args...))
}

func (s *Step) Infos() []string {
	return s.infos
}

func (s *Step) LogError(msg string) {
	s.errs = append(s.errs, msg)
}

func (s *Step) LogErrorf(format string, args ...interface{}) {
	s.LogError(fmt.Sprintf(format, args...))
}

func (s *Step) Errors() []string {
	return s.errs
}

func (s *Step) Prompt(msg string) (string, error) {
	return msg, nil
}

func (s *Step) Reset() {
	s.errs, s.infos, s.status = nil, nil, nil
	s.stopped, s.success = false, false
}
