package dtos

import "github.com/kyma-project/cli.v3/internal/modulesv2/entities"

type CommunityListResult struct {
	Name              string
	Version           string
	ModuleState       string
	InstallationState string
}

func CommunityListResultsFromInstallations(modules []entities.CommunityModuleInstallation) []CommunityListResult {
	results := make([]CommunityListResult, 0, len(modules))
	for _, m := range modules {
		results = append(results, CommunityListResult{
			Name:              m.FullName(),
			Version:           m.Version,
			ModuleState:       m.ModuleState,
			InstallationState: m.InstallationState,
		})
	}
	return results
}
