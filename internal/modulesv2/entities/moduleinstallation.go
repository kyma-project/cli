package entities

import "github.com/kyma-project/cli.v3/internal/kube/kyma"

type ModuleInstallation struct {
	Name                 string
	Version              string
	Channel              string
	ModuleState          string
	Managed              *bool
	CustomResourcePolicy string
	Template             kyma.ModuleStatus
}

func NewModuleInstallationFromRaw(raw kyma.KymaModuleInfo) *ModuleInstallation {
	return &ModuleInstallation{
		Name:                 raw.Status.Name,
		Version:              raw.Status.Version,
		Channel:              raw.Status.Channel,
		ModuleState:          raw.Status.State,
		Managed:              raw.Spec.Managed,
		CustomResourcePolicy: raw.Spec.CustomResourcePolicy,
		Template:             raw.Status,
	}
}

func (m *ModuleInstallation) IsManaged() bool {
	return m.Managed == nil || *m.Managed
}
