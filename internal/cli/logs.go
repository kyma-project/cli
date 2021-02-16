package cli

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
)

//nullLogger does not write any log messages
type nullLogger struct{}

func (nl *nullLogger) Info(args ...interface{})                    {}
func (nl *nullLogger) Infof(template string, args ...interface{})  {}
func (nl *nullLogger) Warn(args ...interface{})                    {}
func (nl *nullLogger) Warnf(template string, args ...interface{})  {}
func (nl *nullLogger) Error(args ...interface{})                   {}
func (nl *nullLogger) Errorf(template string, args ...interface{}) {}
func (nl *nullLogger) Fatal(args ...interface{})                   {}
func (nl *nullLogger) Fatalf(template string, args ...interface{}) {}

//zapLogger is using the ZAP logging API
type zapLogger struct{}

func (l *zapLogger) Info(args ...interface{}) {
	l.Info(args...)
}
func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.Infof(template, args...)
}
func (l *zapLogger) Warn(args ...interface{}) {
	l.Warn(args...)
}
func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.Warnf(template, args...)
}
func (l *zapLogger) Error(args ...interface{}) {
	l.Error(args...)
}
func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.Errorf(template, args...)
}
func (l *zapLogger) Fatal(args ...interface{}) {
	l.Fatal(args...)
}
func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.Fatalf(template, args...)
}

// NewLogger returns the logger used for CLI log output (used in Hydroform deployments)
func NewLogger(printLogs bool) logger.Interface {
	if printLogs {
		return &zapLogger{}
	}
	return &nullLogger{}
}
