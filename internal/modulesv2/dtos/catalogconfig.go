package dtos

const KYMA_COMMUNITY_MODULES_REPOSITORY_URL = "https://kyma-project.github.io/community-modules/all-modules.json"

type CatalogConfig struct {
	ListKyma     bool
	ListCluster  bool
	ExternalUrls []string
}

func NewCatalogConfigFromRemote(remote bool, remoteUrls []string) *CatalogConfig {
	externalUrls := []string{}

	if remote {
		externalUrls = append(externalUrls, KYMA_COMMUNITY_MODULES_REPOSITORY_URL)
	}

	if len(remoteUrls) > 0 {
		externalUrls = append(externalUrls, remoteUrls...)
	}

	if len(externalUrls) > 0 {
		return &CatalogConfig{ExternalUrls: uniqueValues(externalUrls)}
	}

	return &CatalogConfig{ListKyma: true, ListCluster: true}
}

func uniqueValues(urls []string) []string {
	seen := make(map[string]bool)
	dedupedUrls := make([]string, 0, len(urls))
	for _, url := range urls {
		if !seen[url] {
			seen[url] = true
			dedupedUrls = append(dedupedUrls, url)
		}
	}

	return dedupedUrls
}
