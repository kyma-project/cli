package hostfiles

import (
	"context"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/pkg/installation"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultDomain = "kyma.local"
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
		Use:   "add-domain-to-host-file",
		Short: "Adds specified domain to the system host file.",
		Long: `Use this command to add own domain name to the host file of the local system.
`,
		RunE: func(_ *cobra.Command, _ []string) error { return cmd.Run() },
	}

	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", defaultDomain, "Domain to be added")
	cobraCmd.MarkFlagRequired("domain")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if err := cmd.opts.validateFlags(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	if !root.IsWithSudo() {
		return errors.Wrap(err, "Could not add domains to host file. Make sure you are using sudo.")
	}

	clusterConfig, err := installation.GetClusterInfoFromConfigMap(cmd.K8s)
	if err != nil {
		return errors.Wrap(err, "Failed to get cluster information.")
	}

	s := cmd.NewStep("Adding domains to /etc/hosts")
	err = AddDevDomainsToEtcHosts(s, clusterConfig, cmd.K8s, cmd.opts.Domain)
	if err != nil {
		s.Failure()
		if cmd.opts.Verbose {
			s.LogErrorf("error: %v\n", err)
		}
		return err
	}
	s.Successf("Domains added")
	return nil
}

func GetVirtualServiceHostnames(kymaKube kube.KymaKube) ([]string, error) {
	vsList, err := kymaKube.Istio().NetworkingV1alpha3().VirtualServices("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hostnames := []string{}
	for _, v := range vsList.Items {
		hostnames = append(hostnames, v.Spec.Hosts...)
	}

	return hostnames, nil
}

func AddDevDomainsToEtcHosts(
	s step.Step, clusterInfo installation.ClusterInfo, kymaKube kube.KymaKube, domain string) error {
	hostnames := ""

	vsList, err := kymaKube.Istio().NetworkingV1alpha3().VirtualServices("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, v := range vsList.Items {
		for _, host := range v.Spec.Hosts {
			hostnames = hostnames + " " + host
		}
	}

	// if clusterInfo.LocalVMDriver != "none" {
	// 	_, err := minikube.RunCmd(verbose, clusterInfo.Profile, timeout, "ssh", "sudo /bin/sh -c 'echo \""+hostAlias+"\" >> /etc/hosts'")
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	hostAlias := strings.Trim(clusterInfo.LocalIP, "\n") + hostnames
	hostAlias = "127.0.0.1" + hostAlias
	return addDevDomainsToEtcHostsOSSpecific(domain, s, hostAlias)
}
