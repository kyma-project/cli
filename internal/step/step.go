package step

type Step interface {
	Start()
	Status(msg string)
	Success()
	Successf(format string, args ...interface{})
	Failure()
	Failuref(format string, args ...interface{})
	Stop(success bool)
	Stopf(success bool, format string, args ...interface{})
	LogInfo(msg string)
	LogInfof(format string, args ...interface{})
	LogError(msg string)
	LogErrorf(format string, args ...interface{})
	Prompt(msg string) (string, error)
}
