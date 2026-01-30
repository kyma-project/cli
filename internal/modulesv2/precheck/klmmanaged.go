package precheck

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ClusterKLMManaged struct {
	clusterMetadataRepository repository.ClusterMetadataRepository
}

func RunClusterKLMManagedCheck(kymaConfig *cmdcommon.KymaConfig) clierror.Error {
	kubeClient, clierr := kymaConfig.KubeClientConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	check := NewClusterKLMManaged(repository.NewClusterMetadataRepository(kubeClient))
	if !check.Check(kymaConfig.Ctx) {
		return clierror.New(
			"This cluster is not managed by KLM, so the required dependencies (Custom Resource Definitions) are not installed. To install the necessary dependencies, follow these steps:",
			"List available modules from the community modules catalog (official or custom):\n\n\tkyma module catalog --remote\n",
			"Pull a community module to your cluster:\n\n\tkyma module pull <module-name> [--namespace <namespace>]\n",
			"Pulling a community module will install the required dependencies, allowing you to proceed with module installation.",
			"For more information, refer to the documentation or run 'kyma module --help'.")
	}

	return nil
}

func NewClusterKLMManaged(clusterMetadataRepository repository.ClusterMetadataRepository) *ClusterKLMManaged {
	return &ClusterKLMManaged{
		clusterMetadataRepository: clusterMetadataRepository,
	}
}

func (c *ClusterKLMManaged) Check(ctx context.Context) bool {
	return c.clusterMetadataRepository.Get(ctx).IsManagedByKLM
}
