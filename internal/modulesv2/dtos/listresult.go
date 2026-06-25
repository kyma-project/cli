package dtos

import "github.com/kyma-project/cli.v3/internal/modulesv2/entities"

type ListResult struct {
	Name                 string
	Version              string
	Channel              string
	ModuleState          string
	Managed              bool
	CustomResourcePolicy string
	InstallationState    string
}

func ListResultsFromModuleInstallations(modules []entities.ModuleInstallation) []ListResult {
	results := make([]ListResult, 0, len(modules))
	for _, m := range modules {
		results = append(results, ListResult{
			Name:                 m.Name,
			Version:              m.Version,
			Channel:              m.Channel,
			ModuleState:          m.ModuleState,
			Managed:              m.IsManaged(),
			CustomResourcePolicy: m.CustomResourcePolicy,
			InstallationState:    m.InstallationState,
		})
	}
	return results
}
