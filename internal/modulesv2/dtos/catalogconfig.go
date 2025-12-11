package dtos

const KYMA_COMMUNITY_MODULES_REPOSITORY_URL = "https://kyma-project.github.io/community-modules/all-modules.json"

type CatalogConfig struct {
	ListKyma     bool
	ListCluster  bool
	ExternalUrls []string
}

func NewCatalogConfigFromOriginsList(origins []string) *CatalogConfig {
	catalogConfig := &CatalogConfig{}

	for _, origin := range origins {
		switch origin {
		case "kyma":
			catalogConfig.ListKyma = true
		case "cluster":
			catalogConfig.ListCluster = true
		case "community":
			catalogConfig.ExternalUrls = append(catalogConfig.ExternalUrls, KYMA_COMMUNITY_MODULES_REPOSITORY_URL)
		default:
			catalogConfig.ExternalUrls = append(catalogConfig.ExternalUrls, origin)
		}
	}

	return catalogConfig
}
