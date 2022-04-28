package step

var _ Step = mutedStep{}

func NewMutedStep() Step {
	return &mutedStep{}
}

type mutedStep struct {
}

func (m mutedStep) Start() {
}

func (m mutedStep) Status(msg string) {
}

func (m mutedStep) Success() {
}

func (m mutedStep) Successf(format string, args ...interface{}) {
}

func (m mutedStep) Failure() {
}

func (m mutedStep) Failuref(format string, args ...interface{}) {
}

func (m mutedStep) Stop(success bool) {
}

func (m mutedStep) Stopf(success bool, format string, args ...interface{}) {
}

func (m mutedStep) LogInfo(msg string) {
}

func (m mutedStep) LogInfof(format string, args ...interface{}) {
}

func (m mutedStep) LogError(msg string) {
}

func (m mutedStep) LogErrorf(format string, args ...interface{}) {
}

func (m mutedStep) LogWarn(msg string) {
}

func (m mutedStep) LogWarnf(format string, args ...interface{}) {
}

func (m mutedStep) Prompt(msg string) (string, error) {
	return "", nil
}

func (m mutedStep) PromptYesNo(msg string) bool {
	return false
}
