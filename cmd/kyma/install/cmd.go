package install

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/kyma-project/cli/internal/hosts"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/trust"

	"github.com/kyma-project/cli/internal/cli"
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
		
### Description

Before you use the command, make sure your setup meets the following prerequisites:

* Kyma is not installed.
* Kubernetes cluster is available with your kubeconfig file already pointing to it.

Here are the installation steps:

The standard installation uses the minimal configuration. 
Depending on your installation type, the ways to deploy and configure the Kyma Installer are different:

If you install Kyma locally ` + "**from release**" + `, the system does the following:

   1. Fetch the latest or specified release, along with configuration.
   2. Deploy the Kyma Installer on the cluster.
   3. Apply the downloaded or defined configuration.
   
If you install Kyma locally ` + "**from sources**" + `, the system does the following:

   1. Fetch the configuration yaml files from the local sources.
   2. Build the Kyma Installer image.
   3. Deploy the Kyma Installer and apply the fetched configuration.
   
Both installation types continue with the following steps:
   
   4. If overrides have been defined, apply them.
   5. Set the admin password.
   6. Patch the Minikube IP.
   `,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
		Deprecated: "install is deprecated!",
	}

	cobraCmd.Flags().BoolVarP(&o.NoWait, "no-wait", "n", false, "Determines if the command should wait for Kyma installation to complete.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", defaultDomain, "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCert, "tls-cert", "", "", "TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.TLSKey, "tls-key", "", "", "TLS key for the domain used for installation. The key must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", DefaultKymaVersion, `Installation source.
	- To use a specific release, write "kyma install --source=1.15.1".
	- To use the main branch, write "kyma install --source=main".
	- To use a commit, write "kyma install --source=34edf09a".
	- To use a pull request, write "kyma install --source=PR-9486" (only works if '/resources' is modified).
	- To use the local sources, write "kyma install --source=local".
	- To use a custom installer image, write "kyma install --source=user/my-kyma-installer:v1.4.0".`)
	setSource(cobraCmd.Flags().Changed("source"), &o.Source)
	cobraCmd.Flags().StringVarP(&o.LocalSrcPath, "src-path", "", "", "Absolute path to local sources.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 1*time.Hour, "Timeout after which CLI stops watching the installation progress.")
	cobraCmd.Flags().StringVarP(&o.Password, "password", "p", "", "Predefined cluster password.")
	cobraCmd.Flags().StringArrayVarP(&o.OverrideConfigs, "override", "o", nil, "Path to a YAML file with parameters to override.")
	cobraCmd.Flags().StringVarP(&o.ComponentsConfig, "components", "c", "", "Path to a YAML file with a component list to override.")
	cobraCmd.Flags().IntVar(&o.FallbackLevel, "fallback-level", 5, `If "source=main", defines the number of commits from main branch taken into account if artifacts for newer commits do not exist yet`)
	cobraCmd.Flags().StringVarP(&o.CustomImage, "custom-image", "", "", "Full image name including the registry and the tag. Required for installation from local sources to a remote cluster.")
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "", "", "Kyma installation profile (evaluation|production). If not specified, Kyma is installed with the default chart values.")
	return cobraCmd
}

func setSource(isUserDefined bool, source *string) {
	IsRelease, err := strconv.ParseBool(isRelease)
	if err != nil {
		IsRelease = false
		fmt.Println("WARNING: isRelease could not be parsed, continue assuming false value")
	}
	if !isUserDefined && !IsRelease {
		*source = installation.SetKymaSemVersion(*source)
	}
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

	s := cmd.NewStep("Determining cluster type for installation")
	clusterConfig, err := installation.GetClusterInfoFromConfigMap(cmd.K8s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Cluster type determined")

	i, err := cmd.configureInstallation(clusterConfig)
	if err != nil {
		return err
	}

	result, err := i.InstallKyma()
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

	if clusterConfig.IsLocal {
		s = cmd.NewStep("Adding domains to /etc/hosts")
		err = hosts.AddDevDomainsToEtcHosts(s, clusterConfig, cmd.K8s, cmd.opts.Verbose, cmd.opts.Timeout, cmd.opts.Domain)
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

func (cmd *command) configureInstallation(clusterConfig installation.ClusterInfo) (*installation.Installation, error) {

	cmp, err := installation.LoadComponentsConfig(cmd.opts.ComponentsConfig)
	if err != nil {
		return &installation.Installation{}, errors.Wrap(err, "Could not load component configuration file. Make sure file is a valid YAML and contains a component list")
	}
	s, err := installation.NewInstallationService(cmd.K8s.RestConfig(), cmd.opts.Timeout, "", cmp)
	if err != nil {
		return &installation.Installation{}, errors.Wrap(err, "Failed to create installation service. Make sure your kubeconfig is valid")
	}

	return &installation.Installation{
		K8s:     cmd.K8s,
		Service: s,
		Options: &installation.Options{
			NoWait:           cmd.opts.NoWait,
			Verbose:          cmd.opts.Verbose,
			CI:               cmd.opts.CI,
			NonInteractive:   cmd.Factory.NonInteractive,
			Timeout:          cmd.opts.Timeout,
			CustomImage:      cmd.opts.CustomImage,
			Domain:           cmd.opts.Domain,
			TLSCert:          cmd.opts.TLSCert,
			TLSKey:           cmd.opts.TLSKey,
			LocalSrcPath:     cmd.opts.LocalSrcPath,
			Password:         cmd.opts.Password,
			OverrideConfigs:  cmd.opts.OverrideConfigs,
			ComponentsConfig: cmd.opts.ComponentsConfig,
			Source:           cmd.opts.Source,
			FallbackLevel:    cmd.opts.FallbackLevel,
			Profile:          cmd.opts.Profile,
			IsLocal:          clusterConfig.IsLocal,
			LocalCluster: &installation.LocalCluster{
				IP:       clusterConfig.LocalIP,
				Profile:  clusterConfig.Profile,
				Provider: clusterConfig.Provider,
				VMDriver: clusterConfig.LocalVMDriver,
			},
		},
	}, nil
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

func (cmd *command) printSummary(result *installation.Result) error {
	nicePrint := nice.Nice{
		NonInteractive: cmd.Factory.NonInteractive,
	}

	fmt.Println()
	nicePrint.PrintKyma()
	fmt.Print(" is installed in version:\t")
	nicePrint.PrintImportant(result.KymaVersion)

	nicePrint.PrintKyma()
	fmt.Print(" installation took:\t\t")
	nicePrint.PrintImportantf("%d hours %d minutes",
		int64(result.Duration.Hours()), int64(result.Duration.Minutes()))

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
