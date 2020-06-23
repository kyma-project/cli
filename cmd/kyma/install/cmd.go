package install

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-project/cli/pkg/step"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/trust"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"

	"github.com/kyma-project/cli/internal/minikube"
	"github.com/kyma-project/cli/pkg/installation"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

const (
	defaultDomain = "kyma.local"
)

type command struct {
	opts *Options
	cli.Command
}

type clusterInfo struct {
	isLocal       bool
	provider      string
	profile       string
	localIP       string
	localVMDriver string
}

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "install",
		Short: "Installs Kyma on a running Kubernetes cluster.",
		Long: `Use this command to install Kyma on a running Kubernetes cluster.

### Detailed description

Before you use the command, make sure your setup meets the following prerequisites:

* Kyma is not installed.
* Kubernetes cluster is available with your kubeconfig file already pointing to it.

Here are the installation steps:

The standard installation uses the minimal configuration. The system performs the following steps:
1. Deploys and configures the Kyma Installer. At this point, steps differ depending on the installation type.

    When you install Kyma locally ` + "**from release**" + `, the system:
    1. Fetches the latest or specified release along with configuration.
    2. Deploys the Kyma Installer on the cluster.
    3. Applies downloaded or defined configuration.
    4. Applies overrides, if applicable.
    5. Sets the admin password.
    6. Patches the Minikube IP.
	
    When you install Kyma locally ` + "**from sources**" + `, the system:
    1. Fetches the configuration yaml files from the local sources.
    2. Builds the Kyma Installer image.
    3. Deploys the Kyma Installer and applies the fetched configuration.
    4. Applies overrides, if applicable.
    5. Sets the admin password.
    6. Patches the Minikube IP.
    
2. Runs Kyma installation until the ` + "**installed**" + ` status confirms the successful installation. You can override the standard installation settings using the ` + "`--override`" + ` flag.

`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Flag that determines if the command should wait for Kyma installation to complete.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", defaultDomain, "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCert, "tlsCert", "", "", "TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.TLSKey, "tlsKey", "", "", "TLS key for the domain used for installation. The key must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", DefaultKymaVersion, `Installation source. 
	- To use the specific release, write "kyma install --source=1.3.0".
	- To use the latest master, write "kyma install --source=latest".
	- To use the latest published master, which is the latest commit with released images, write "kyma install --source=latest-published".
	- To use a commit, write "kyma install --source=34edf09a". 
	- To use the local sources, write "kyma install --source=local". 
	- To use a custom installer image, write kyma "install --source=user/my-kyma-installer:v1.4.0".`)
	cobraCmd.Flags().StringVarP(&o.LocalSrcPath, "src-path", "", "", "Absolute path to local sources.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 1*time.Hour, "Time-out after which CLI stops watching the installation progress.")
	cobraCmd.Flags().StringVarP(&o.Password, "password", "p", "", "Predefined cluster password.")
	cobraCmd.Flags().StringArrayVarP(&o.OverrideConfigs, "override", "o", nil, "Path to a YAML file with parameters to override.")
	cobraCmd.Flags().StringVarP(&o.ComponentsConfig, "components", "c", "", "Path to a YAML file with component list to override.")
	cobraCmd.Flags().IntVar(&o.FallbackLevel, "fallbackLevel", 5, `If "source=latest-published", defines the number of commits from master branch taken into account if artifacts for newer commits do not exist yet`)
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	var err error
	if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	s := cmd.NewStep("Reading cluster info from ConfigMap")
	clusterConfig, err := cmd.getClusterInfoFromConfigMap()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Cluster info read")

	installation := cmd.configureInstallation(clusterConfig)
	result, err := installation.InstallKyma()
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	if !cmd.opts.CI {
		if err := cmd.importCertificate(trust.NewCertifier(cmd.K8s)); err != nil {
			// certificate import errors do not mean installation failed
			cmd.CurrentStep.LogError(err.Error())
		}
	}

	if clusterConfig.isLocal {
		s = cmd.NewStep("Adding domains to /etc/hosts")
		err = cmd.addDevDomainsToEtcHosts(s, clusterConfig)
		if err != nil {
			s.Failure()
			return err
		}
		s.Successf("Domains added")
	}

	err = cmd.printSummary(result)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) configureInstallation(clusterConfig clusterInfo) *installation.Installation {
	return &installation.Installation{
		Options: &installation.Options{
			NoWait:           cmd.opts.NoWait,
			Verbose:          cmd.opts.Verbose,
			CI:               cmd.opts.CI,
			NonInteractive:   cmd.Factory.NonInteractive,
			Timeout:          cmd.opts.Timeout,
			KubeconfigPath:   cmd.opts.KubeconfigPath,
			Domain:           cmd.opts.Domain,
			TLSCert:          cmd.opts.TLSCert,
			TLSKey:           cmd.opts.TLSKey,
			LocalSrcPath:     cmd.opts.LocalSrcPath,
			Password:         cmd.opts.Password,
			OverrideConfigs:  cmd.opts.OverrideConfigs,
			ComponentsConfig: cmd.opts.ComponentsConfig,
			Source:           cmd.opts.Source,
			FallbackLevel:    cmd.opts.FallbackLevel,
			IsLocal:          clusterConfig.isLocal,
			LocalCluster: &installation.LocalCluster{
				IP:       clusterConfig.localIP,
				Profile:  clusterConfig.profile,
				Provider: clusterConfig.provider,
				VMDriver: clusterConfig.localVMDriver,
			},
		},
	}
}

func (cmd *command) importCertificate(ca trust.Certifier) error {
	if !cmd.opts.NoWait {
		// get cert from cluster
		cert, err := ca.Certificate()
		if err != nil {
			return err
		}

		tmpFile, err := ioutil.TempFile(os.TempDir(), "kyma-*.crt")
		if err != nil {
			return errors.Wrap(err, "Cannot create temporary file for Kyma certificate")
		}
		defer os.Remove(tmpFile.Name())

		if _, err = tmpFile.Write(cert); err != nil {
			return errors.Wrap(err, "Failed to write the kyma certificate")
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}

		if err := ca.StoreCertificate(tmpFile.Name(), cmd.CurrentStep); err != nil {
			return err
		}
		cmd.CurrentStep.Successf("Kyma root certificate imported")

	} else {
		cmd.CurrentStep.LogError(ca.Instructions())
	}
	return nil
}

func (cmd *command) addDevDomainsToEtcHosts(s step.Step, clusterInfo clusterInfo) error {
	hostnames := ""

	vsList, err := cmd.K8s.Istio().NetworkingV1alpha3().VirtualServices("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, v := range vsList.Items {
		for _, host := range v.Spec.Hosts {
			hostnames = hostnames + " " + host
		}
	}

	hostAlias := "127.0.0.1" + hostnames

	if clusterInfo.localVMDriver != "none" {
		_, err := minikube.RunCmd(cmd.opts.Verbose, clusterInfo.profile, cmd.opts.Timeout, "ssh", "sudo /bin/sh -c 'echo \""+hostAlias+"\" >> /etc/hosts'")
		if err != nil {
			return err
		}
	}

	hostAlias = strings.Trim(clusterInfo.localIP, "\n") + hostnames

	return addDevDomainsToEtcHostsOSSpecific(cmd.opts.Domain, s, hostAlias)
}

func (cmd *command) getClusterInfoFromConfigMap() (clusterInfo, error) {
	cm, err := cmd.K8s.Static().CoreV1().ConfigMaps("kube-system").Get("kyma-cluster-info", metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return clusterInfo{}, nil
		}
		return clusterInfo{}, err
	}

	isLocal, err := strconv.ParseBool(cm.Data["isLocal"])
	if err != nil {
		isLocal = false
	}

	clusterConfig := clusterInfo{
		isLocal:       isLocal,
		provider:      cm.Data["provider"],
		profile:       cm.Data["profile"],
		localIP:       cm.Data["localIP"],
		localVMDriver: cm.Data["localVMDriver"],
	}

	return clusterConfig, nil
}

func (cmd *command) printSummary(result *installation.Result) error {
	nicePrint := nice.Nice{}
	if cmd.Factory.NonInteractive {
		nicePrint.NonInteractive = true
	}

	fmt.Println()
	nicePrint.PrintKyma()
	fmt.Print(" is installed in version:\t")
	nicePrint.PrintImportant(result.KymaVersion)

	nicePrint.PrintKyma()
	fmt.Print(" is running at:\t\t")
	nicePrint.PrintImportant(result.Host)

	nicePrint.PrintKyma()
	fmt.Print(" console:\t\t\t")
	nicePrint.PrintImportantf(result.Console)

	nicePrint.PrintKyma()
	fmt.Print(" admin email:\t\t")
	nicePrint.PrintImportant(result.AdminEmail)

	if cmd.opts.Password == "" && !cmd.Factory.NonInteractive {
		nicePrint.PrintKyma()
		fmt.Printf(" admin password:\t\t")
		nicePrint.PrintImportant(result.AdminPassword)
	}

	for _, warning := range result.Warnings {
		nicePrint.PrintImportant(warning)
	}

	fmt.Printf("\nHappy ")
	nicePrint.PrintKyma()
	fmt.Printf("-ing! :)\n\n")

	return nil
}
