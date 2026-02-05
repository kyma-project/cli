package precheck

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

// -----------------------------------------------------------------------------
// RequireKLMManaged - Guard check that fails if cluster is not managed by KLM
// -----------------------------------------------------------------------------

// KLMRequirer checks if the cluster is managed by Kyma Lifecycle Manager (KLM).
type KLMRequirer struct {
	clusterMetadataRepository repository.ClusterMetadataRepository
}

// NewKLMRequirer creates a new KLMRequirer instance.
func NewKLMRequirer(client kube.Client) *KLMRequirer {
	return &KLMRequirer{
		clusterMetadataRepository: repository.NewClusterMetadataRepository(client),
	}
}

// RequireKLMManaged verifies that the cluster is managed by Kyma Lifecycle Manager (KLM).
// Returns an error if the cluster is not KLM-managed, guiding the user on how to proceed.
func RequireKLMManaged(kymaConfig *cmdcommon.KymaConfig, cmdGroup string) clierror.Error {
	kubeClient, clierr := kymaConfig.KubeClientConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	requirer := NewKLMRequirer(kubeClient)
	if !requirer.IsKLMManaged(kymaConfig.Ctx) {
		return clierror.New(
			"This cluster is not managed by Kyma Lifecycle Manager (KLM).",
			"This command requires a KLM-managed cluster and cannot be used with community modules pulled manually.",
			fmt.Sprintf("For more information, refer to the documentation or run '%s module --help'.", cmdGroup),
		)
	}

	return nil
}

// IsKLMManaged checks if the cluster is managed by Kyma Lifecycle Manager.
func (r *KLMRequirer) IsKLMManaged(ctx context.Context) bool {
	return r.clusterMetadataRepository.Get(ctx).IsManagedByKLM
}
