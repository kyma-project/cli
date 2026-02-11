package precheck

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Command group prefixes for user-facing messages
const (
	CmdGroupStable = "kyma"
	CmdGroupAlpha  = "kyma alpha"
)

// CRDRequirer checks if the ModuleTemplate CRD is installed on the cluster.
type CRDRequirer struct {
	client kube.Client
}

// NewCRDRequirer creates a new CRDRequirer instance.
func NewCRDRequirer(client kube.Client) *CRDRequirer {
	return &CRDRequirer{
		client: client,
	}
}

// RequireCRD verifies that the ModuleTemplate CRD is installed on the cluster.
// Returns an error if the CRD is missing, guiding the user to install it.
func RequireCRD(kymaConfig *cmdcommon.KymaConfig, cmdGroup string) clierror.Error {
	kubeClient, clierr := kymaConfig.KubeClientConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	requirer := NewCRDRequirer(kubeClient)
	if !requirer.IsInstalled(kymaConfig.Ctx) {
		return clierror.New(
			"This cluster is not managed by KLM, so the required dependencies (Custom Resource Definitions) are not installed.",
			fmt.Sprintf("List available modules from the community modules catalog (official or custom):\n\t$ %s module catalog --remote", cmdGroup),
			fmt.Sprintf("Pull a community module to your cluster:\n\t$ %s module pull <module-name> [--namespace <namespace>]", cmdGroup),
			"Pulling a community module will install the required dependencies, allowing you to proceed with module installation.",
			fmt.Sprintf("For more information, refer to the documentation or run '%s module --help'.", cmdGroup),
		)
	}

	return nil
}

// IsInstalled checks if the ModuleTemplate CRD exists on the cluster.
func (r *CRDRequirer) IsInstalled(ctx context.Context) bool {
	_, err := r.fetchStoredCRD(ctx)
	return err == nil
}

func (r *CRDRequirer) fetchStoredCRD(ctx context.Context) (*unstructured.Unstructured, error) {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion(moduleTemplateCRDMeta.APIVersion)
	crd.SetKind(moduleTemplateCRDMeta.Kind)
	crd.SetName(moduleTemplateCRDMeta.Name)

	return r.client.RootlessDynamic().Get(ctx, crd)
}
