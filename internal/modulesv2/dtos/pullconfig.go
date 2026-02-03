package dtos

type PullConfig struct {
	Namespace           string
	ModuleName          string
	Version             string
	RemoteRepositoryUrl string
}

func NewPullConfig(moduleName, namespace, remote, version string) *PullConfig {
	var remoteRepo string

	if remote == "" {
		remoteRepo = KYMA_COMMUNITY_MODULES_REPOSITORY_URL
	} else {
		remoteRepo = remote
	}

	return &PullConfig{
		ModuleName:          moduleName,
		Namespace:           namespace,
		Version:             version,
		RemoteRepositoryUrl: remoteRepo,
	}
}
