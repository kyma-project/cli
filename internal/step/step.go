package step

import "fmt"

type Step interface {
	Start()
	Status(msg string)
	Success()
	Successf(format string, args ...interface{})
	Failure()
	Failuref(format string, args ...interface{})
	Stop(success bool)
	Stopf(success bool, format string, args ...interface{})
}

const (
	successGliph = "✅"
	failureGliph = "❌"
)

func NewSimpleStep(msg string) Step {
	return &simpleStep{msg}
}

type simpleStep struct {
	msg string
}

func (s *simpleStep) Start() {
	fmt.Println(s.msg)
}

func (s *simpleStep) Status(msg string) {
	fmt.Printf("%s : %s\n", s.msg, msg)
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
	var gliph string
	if success {
		gliph = successGliph
	} else {
		gliph = failureGliph
	}
	fmt.Printf("%s %s\n", s.msg, gliph)
}
