package kubectl

import "fmt"

type Wrapper struct {
	verbose    bool
	kubeconfig string
}

func NewWrapper(verbose bool, kubeconfig string) *Wrapper {
	return &Wrapper{
		verbose:    verbose,
		kubeconfig: kubeconfig,
	}
}

func (w *Wrapper) RunCmd(args ...string) (string, error) {
	args = append(args, fmt.Sprintf("--kubeconfig=%s", w.kubeconfig))
	return runCmd(w.verbose, args...)
}

func (w *Wrapper) RunApplyCmd(resources []map[string]interface{}) (string, error) {
	return runApplyCmd(resources, w.verbose, w.kubeconfig)
}
