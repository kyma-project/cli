package cluster

import (
	"fmt"
	"regexp"

	"strconv"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	MINIKUBE_VERSION string = "0.28.2"
	KUBECTL_VERSION         = "1.10.0"
)

type MinikubeOptions struct {
	Domain   string
	VMDriver string
}

func NewMinikubeOptions() *MinikubeOptions {
	return &MinikubeOptions{}
}

func NewMinikubeCmd(o *MinikubeOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "minikube",
		Short: "Prepares a minikube cluster",
		Long: `Prepares a minikube cluster
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use")
	cmd.Flags().StringVarP(&o.VMDriver, "vm-driver", "", "hyperkit", "VMDriver to use")
	return cmd
}

func (o *MinikubeOptions) Run() error {
	fmt.Printf("Installing minikube cluster using domain '%s' and vm-driver '%s'\n", o.Domain, o.VMDriver)

	err := checkMinikubeVersion()
	if err != nil {
		return err
	}

	err = checkKubectlVersion()
	if err != nil {
		return err
	}

	return nil
}

func checkMinikubeVersion() error {

	versionCmd := []string{"version"}
	versionText := internal.RunMinikubeCmd(versionCmd)

	exp, _ := regexp.Compile("minikube version: v((\\d+.\\d+.\\d+))")
	version := exp.FindStringSubmatch(versionText)

	if version[1] != MINIKUBE_VERSION {
		return fmt.Errorf("Currently minikube in version '%s' is required", MINIKUBE_VERSION)
	}
	return nil

}

func checkKubectlVersion() error {
	versionCmd := []string{"version"}
	versionText := internal.RunKubeCmd(versionCmd)

	exp, _ := regexp.Compile("Client Version: v((\\d+).(\\d+).(\\d+))")
	kubctlIsVersion := exp.FindStringSubmatch(versionText)

	exp, _ = regexp.Compile("((\\d+).(\\d+).(\\d+))")
	kubctlMustVersion := exp.FindStringSubmatch(KUBECTL_VERSION)

	majorIsVersion, _ := strconv.Atoi(kubctlIsVersion[2])
	majorMustVersion, _ := strconv.Atoi(kubctlMustVersion[2])
	minorIsVersion, _ := strconv.Atoi(kubctlIsVersion[3])
	minorMustVersion, _ := strconv.Atoi(kubctlMustVersion[3])

	if minorIsVersion-minorMustVersion < 1 || minorIsVersion-minorMustVersion > 1 {
		fmt.Errorf("Your kubectl version is '%s'. Supported versions of kubectl are from '%s.%s.*' to '%s.%s.*'", kubctlIsVersion, majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	if majorIsVersion != majorMustVersion {
		fmt.Errorf("Your kubectl version is '%s'. Supported versions of kubectl are from '%s.%s.*' to '%s.%s.*'", kubctlIsVersion, majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	return nil
}
