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
		return &CatalogConfig{ExternalUrls: remote.Values}
	}

	return &CatalogConfig{ListKyma: true, ListCluster: true}
}
