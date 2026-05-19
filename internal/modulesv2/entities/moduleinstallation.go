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
	}
}

func (m *ModuleInstallation) IsManaged() bool {
	return m.Managed == nil || *m.Managed
}
