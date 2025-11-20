package dtos

import "github.com/kyma-project/cli.v3/internal/modulesv2/entities"

const KYMA_ORIGIN = "kyma"
const COMMUNITY_ORIGIN = "community"

type CatalogResult struct {
	Name              string
	AvailableVersions []string
	Origin            string
}

func CatalogResultFromCoreModuleTemplates(coreModuleTemplates []entities.CoreModuleTemplate) []CatalogResult {
	results := []CatalogResult{}

	// Cache to quickly get an index of module that's already present in the result set
	resultsCache := map[string]int{}

	for _, coreModuleTemplate := range coreModuleTemplates {
		if i, exists := resultsCache[coreModuleTemplate.ModuleName]; exists {
			results[i].AvailableVersions = append(results[i].AvailableVersions, coreModuleTemplate.GetVersionWithChannel())
		} else {
			newResult := CatalogResult{
				Name:              coreModuleTemplate.ModuleName,
				AvailableVersions: []string{coreModuleTemplate.GetVersionWithChannel()},
				Origin:            KYMA_ORIGIN,
			}
			results = append(results, newResult)
			resultsCache[coreModuleTemplate.ModuleName] = len(results) - 1
		}
	}

	return results
}

func CatalogResultFromCommunityModuleTemplates(communityModuleTemplates []entities.CommunityModuleTemplate) []CatalogResult {
	results := []CatalogResult{}

	// Cache key: moduleName + origin
	resultsCache := map[string]int{}

	for _, communityModuleTemplate := range communityModuleTemplates {
		origin := getOriginFor(communityModuleTemplate)
		cacheKey := communityModuleTemplate.ModuleName + "|" + origin

		if i, exists := resultsCache[cacheKey]; exists {
			results[i].AvailableVersions = append(results[i].AvailableVersions, communityModuleTemplate.Version)
		} else {
			newResult := CatalogResult{
				Name:              communityModuleTemplate.ModuleName,
				AvailableVersions: []string{communityModuleTemplate.Version},
				Origin:            origin,
			}
			results = append(results, newResult)
			resultsCache[cacheKey] = len(results) - 1
		}
	}

	return results
}

func getOriginFor(communityModuleTemplate entities.CommunityModuleTemplate) string {
	if communityModuleTemplate.IsExternal() {
		return COMMUNITY_ORIGIN
	}

	return communityModuleTemplate.GetNamespacedName()
}
