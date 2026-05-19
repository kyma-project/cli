package entities

import "github.com/kyma-project/cli.v3/internal/kube/kyma"

type ModuleInstallation struct {
	Name                 string
	Version              string
	Channel              string
	ModuleState          string
	Managed              *bool
	CustomResourcePolicy string
	TemplateName         string
	TemplateNamespace    string
	specModuleName       string
	statusModuleName     string
}

func NewModuleInstallationFromRaw(raw kyma.KymaModuleInfo) *ModuleInstallation {
	name := raw.Status.Name
	if name == "" {
		name = raw.Spec.Name
	}
	return &ModuleInstallation{
		Name:                 name,
		Version:              raw.Status.Version,
		Channel:              raw.Status.Channel,
		ModuleState:          raw.Status.State,
		Managed:              raw.Spec.Managed,
		CustomResourcePolicy: raw.Spec.CustomResourcePolicy,
		TemplateName:         raw.Status.Template.GetName(),
		TemplateNamespace:    raw.Status.Template.GetNamespace(),
		specModuleName:       raw.Spec.Name,
		statusModuleName:     raw.Status.Name,
	}
}

func (m *ModuleInstallation) IsManaged() bool {
	return m.Managed == nil || *m.Managed
}

func (m *ModuleInstallation) IsBeingDeleted() bool {
	return m.statusModuleName != "" && m.specModuleName == ""
}
