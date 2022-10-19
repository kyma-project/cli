package provision

import (
	"github.com/spf13/cobra"
)

const DefaultK8sShortVersion = "1.24"                       //default Kubernetes version for provisioning clusters on hyperscalers
const DefaultK8sFullVersion = DefaultK8sShortVersion + ".6" //default Kubernetes version with the "patch" component (mainly for K3d/K3s)

// NewCmd creates a new provision command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a cluster for Kyma installation.",
	}
	return cmd
}
