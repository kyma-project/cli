package dtos

type PullResult struct {
	ModuleName         string
	ModuleTemplateName string
	Version            string
	Namespace          string
}

func NewPullResult(moduleName, templateName, version, namespace string) *PullResult {
	return &PullResult{
		ModuleName:         moduleName,
		ModuleTemplateName: templateName,
		Version:            version,
		Namespace:          namespace,
	}
}
