package version

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

type command struct {
	opts *Options
	cli.Command
}

//Version contains the cli binary version injected by the build system
var Version string

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version of Kyma CLI and of the connected Kyma cluster.",
		Long: `Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"v"},
	}

	cobraCmd.Flags().BoolVarP(&o.ClientOnly, "client", "c", false, "Client version only (no server required)")
	cobraCmd.Flags().BoolVarP(&o.VersionDetails, "details", "d", false, "Detailed information for each Kyma version")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var w io.Writer = os.Stdout

	fmt.Fprintf(w, "Kyma CLI version: %s\n", versionOrDefault(Version))

	if cmd.opts.ClientOnly {
		//we are done
		return nil
	}

	//print Kyma Version
	err := cmd.setKubeClient()
	if err != nil {
		return err
	}

	isKyma2, err := checkKyma2(cmd.K8s)
	if err != nil {
		return err
	}
	if isKyma2 {
		//Check for kyma 2
		version, err := getKyma2Version(cmd.K8s)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Kyma 2 cluster versions: %s\n", versionOrDefault(version))

	} else {
		// Print kyma 1 version
		version, err := getKyma1Version(cmd.K8s)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Kyma 1 cluster versions: %s\n", versionOrDefault(version))
	}

	return nil
}

func (cmd *command) setKubeClient() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return  errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	return nil
}

func versionOrDefault(version string) string {
	return stringOrDefault(version, "N/A")
}

func stringOrDefault(s, def string) string {
	if len(s) == 0 {
		return def
	}
	return s
}

func getDeployments(k8s kube.KymaKube) (*v1.DeploymentList , error) {
	return  k8s.Static().AppsV1().Deployments("kyma-system").List(context.Background(), metav1.ListOptions{LabelSelector: "reconciler.kyma-project.io/managed-by=reconciler"})
}

func getKyma2Version(k8s kube.KymaKube) (string, error) {
	deps, err := getDeployments(k8s)
	if err != nil{
		return "N/A", err
	}
	if len(deps.Items) == 0 {
		return "N/A", nil
	}
	return deps.Items[0].Labels["reconciler.kyma-project.io/origin-version"], nil
}

func getKyma1Version(k8s kube.KymaKube) (string, error) {
	pods, err := k8s.Static().CoreV1().Pods("kyma-installer").List(context.Background(), metav1.ListOptions{LabelSelector: "name=kyma-installer"})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "N/A", nil
	}

	imageParts := strings.Split(pods.Items[0].Spec.Containers[0].Image, ":")
	if len(imageParts) < 2 {
		return "N/A", nil
	}

	return imageParts[1], nil
}

func checkKyma2(k8s kube.KymaKube) (bool, error) {
	deps, err := getDeployments(k8s)
	if err != nil{
		return false, err
	}
	if len(deps.Items) == 0 {
		return false, nil
	}
	return true, nil
}

//KymaVersion determines the version of kyma installed in the cluster succesible via the provided kubernetes client
func KymaVersion(k8s kube.KymaKube) (string, error) {
	isKyma2, err := checkKyma2(k8s)
	if err != nil {
		return "", err
	}
	if isKyma2 {
		//Check for kyma 2
		return getKyma2Version(k8s)
	} else {
		return getKyma1Version(k8s)
	}
}
