package entities

import "fmt"

type CommunityModuleInstallation struct {
	Name              string
	Namespace         string
	Version           string
	ModuleState       string
	InstallationState string
}

func (m *CommunityModuleInstallation) FullName() string {
	return fmt.Sprintf("%s/%s", m.Namespace, m.Name)
}
