package step

var _ Step = mutedStep{}

func NewMutedStep() Step {
	return &mutedStep{}
}

type mutedStep struct {
}

func (m mutedStep) Start() {
}

func (m mutedStep) Status(_ string) {
}

func (m mutedStep) Success() {
}

func (m mutedStep) Successf(_ string, _ ...interface{}) {
}

func (m mutedStep) Failure() {
}

func (m mutedStep) Failuref(_ string, _ ...interface{}) {
}

func (m mutedStep) Stop(_ bool) {
}

func (m mutedStep) Stopf(_ bool, _ string, _ ...interface{}) {
}

func (m mutedStep) LogInfo(_ string) {
}

func (m mutedStep) LogInfof(_ string, _ ...interface{}) {
}

func (m mutedStep) LogError(_ string) {
}

func (m mutedStep) LogErrorf(_ string, _ ...interface{}) {
}

func (m mutedStep) LogWarn(_ string) {
}

func (m mutedStep) LogWarnf(_ string, _ ...interface{}) {
}

func (m mutedStep) Prompt(_ string) (string, error) {
	return "", nil
}

func (m mutedStep) PromptYesNo(_ string) bool {
	return false
}
