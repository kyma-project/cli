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

func (w *Wrapper) WaitForPodReady(namespace string, labelName string, labelValue string) error {
	return waitForPodReady(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) WaitForPodGone(namespace string, labelName string, labelValue string) error {
	return waitForPodGone(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsPodDeployed(namespace string, labelName string, labelValue string) (bool, error) {
	return isPodDeployed(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsResourceDeployed(resource string, namespace string, labelName string, labelValue string) (bool, error) {
	return isResourceDeployed(resource, namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsClusterResourceDeployed(resource string, labelName string, labelValue string) (bool, error) {
	return isClusterResourceDeployed(resource, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsPodReady(namespace string, labelName string, labelValue string) (bool, error) {
	return isPodReady(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) CheckVersion() (string, error) {
	return checkVersion(w.verbose)
}
