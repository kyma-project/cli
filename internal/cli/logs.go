package cli

import (
	"fmt"
	"log"

	"github.com/go-logr/zapr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// NewLogger returns the logger used for CLI log output (used in Hydroform deployments)
func NewLogger(printLogs bool) *zap.Logger {
	var logger *zap.Logger
	if printLogs {
		var err error

		if logger, err = createVerboseLogger(); err != nil {
			log.Fatalf("Can't initialize zap logger: %v", err)
		}

	} else {
		logger = zap.NewNop()
	}
	logr := zapr.NewLoggerWithOptions(logger)
	ctrllog.SetLogger(zapr.NewLoggerWithOptions(zap.NewNop()))
	klog.SetLogger(logr)
	ocm.DefaultContext().LoggingContext().SetBaseLogger(logr)
	ocm.DefaultContext().LoggingContext().SetDefaultLevel(9)
	ocm.DefaultContext().CredentialsContext().LoggingContext().SetBaseLogger(logr)
	ocm.DefaultContext().CredentialsContext().LoggingContext().SetDefaultLevel(9)

	return logger
}

func createVerboseLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(-9))
	config.DisableStacktrace = true
	return config.Build()
}

// NewHydroformLoggerAdapter adapts a ZAP logger to a Hydrofrom compatible logger
func NewHydroformLoggerAdapter(logger *zap.Logger) *HydroformLoggerAdapter {
	return &HydroformLoggerAdapter{
		logger: logger,
	}
}

// HydroformLoggerAdapter is implementing the logger interface of Hydroform
// to make it compatible with the ZAP logger API.
type HydroformLoggerAdapter struct {
	logger *zap.Logger
}

func (l *HydroformLoggerAdapter) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprintf("%v", args))
}
func (l *HydroformLoggerAdapter) Infof(template string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(template, args...))
}
func (l *HydroformLoggerAdapter) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprintf("%v", args...))
}
func (l *HydroformLoggerAdapter) Warnf(template string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(template, args...))
}
func (l *HydroformLoggerAdapter) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprintf("%v", args))
}
func (l *HydroformLoggerAdapter) Errorf(template string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(template, args...))
}
func (l *HydroformLoggerAdapter) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf("%v", args))
}
func (l *HydroformLoggerAdapter) Fatalf(template string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(template, args...))
}
