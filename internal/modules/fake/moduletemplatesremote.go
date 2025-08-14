package fake

import "github.com/kyma-project/cli.v3/internal/kube/kyma"

type ModuleTemplatesRemoteRepo struct {
	ReturnCommunity []kyma.ModuleTemplate
	CommunityErr    error
}

func (r *ModuleTemplatesRemoteRepo) Community() ([]kyma.ModuleTemplate, error) {
	return r.ReturnCommunity, r.CommunityErr
}
