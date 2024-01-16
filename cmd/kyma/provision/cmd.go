package provision

import (
	"github.com/spf13/cobra"
)

const DefaultK8sShortVersion = "1.27"                       // default Kubernetes version for provisioning clusters on hyperscalers
const DefaultK8sFullVersion = DefaultK8sShortVersion + ".9" // default Kubernetes version with the "patch" component (mainly for K3d/K3s)
const DefaultGardenLinuxVersion = "934.9.0"                 // default Garden Linux version

// NewCmd creates a new provision command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a cluster for Kyma installation.",
	}
	return cmd
}
