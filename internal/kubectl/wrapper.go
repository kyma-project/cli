package kubectl

type Wrapper struct {
	verbose bool
}

func NewWrapper(verbose bool) *Wrapper {
	return &Wrapper{verbose}
}

func (w *Wrapper) RunCmd(args ...string) (string, error) {
	return RunCmd(w.verbose, args...)
}

func (w *Wrapper) RunApplyCmd(resources []map[string]interface{}) (string, error) {
	return RunApplyCmd(resources, w.verbose)
}

func (w *Wrapper) WaitForPodReady(namespace string, labelName string, labelValue string) error {
	return WaitForPodReady(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) WaitForPodGone(namespace string, labelName string, labelValue string) error {
	return WaitForPodGone(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsPodDeployed(namespace string, labelName string, labelValue string) (bool, error) {
	return IsPodDeployed(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsResourceDeployed(resource string, namespace string, labelName string, labelValue string) (bool, error) {
	return IsResourceDeployed(resource, namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsClusterResourceDeployed(resource string, labelName string, labelValue string) (bool, error) {
	return IsClusterResourceDeployed(resource, labelName, labelValue, w.verbose)
}

func (w *Wrapper) IsPodReady(namespace string, labelName string, labelValue string) (bool, error) {
	return IsPodReady(namespace, labelName, labelValue, w.verbose)
}

func (w *Wrapper) CheckVersion() (string, error) {
	return CheckVersion(w.verbose)
}

