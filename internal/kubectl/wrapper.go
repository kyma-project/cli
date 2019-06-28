package kubectl

type Wrapper struct {
	verbose bool
}

func NewWrapper(verbose bool) *Wrapper {
	return &Wrapper{verbose}
}

func (w *Wrapper) RunCmd(args ...string) (string, error) {
	return runCmd(w.verbose, args...)
}

func (w *Wrapper) RunApplyCmd(resources []map[string]interface{}) (string, error) {
	return runApplyCmd(resources, w.verbose)
}
