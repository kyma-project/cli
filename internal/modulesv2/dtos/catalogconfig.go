package dtos

import (
	"github.com/kyma-project/cli.v3/internal/flags"
)

const KYMA_COMMUNITY_MODULES_REPOSITORY_URL = "https://kyma-project.github.io/community-modules/all-modules.json"

type CatalogConfig struct {
	ListKyma     bool
	ListCluster  bool
	ExternalUrls []string
}

func NewCatalogConfigFromRemote(remote flags.BoolOrStrings) *CatalogConfig {
	if remote.Enabled && len(remote.Values) == 0 {
		return &CatalogConfig{ExternalUrls: []string{KYMA_COMMUNITY_MODULES_REPOSITORY_URL}}
	}

	if remote.Enabled && len(remote.Values) > 0 {
		return &CatalogConfig{ExternalUrls: uniqueValues(remote.Values)}
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
