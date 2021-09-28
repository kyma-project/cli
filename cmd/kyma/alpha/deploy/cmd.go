package deploy

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/logger"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/component"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/k3d"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	//Register all reconcilers
	_ "github.com/kyma-incubator/reconciler/pkg/reconciler/instances"
)

const defaultVersion = "main"
const istioChartPath = "/resources/istio-configuration/Chart.yaml"

type command struct {
	cli.Command
	opts *Options
}

//NewCmd creates a new deploy command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run(cmd.opts) },
		Aliases: []string{"d"},
	}
	cobraCmd.Flags().StringSliceVarP(&o.Components, "component", "", []string{}, "Provide one or more components to deploy (e.g. --component componentName@namespace)")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components-file", "c", "", `Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")`)
	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", "", `Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")`)
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", defaultVersion, `Installation source:
	- Deploy a specific release, for example: "kyma deploy --source=2.0.0"
	- Deploy a specific branch of the Kyma repository on kyma-project.org: "kyma deploy --source=<my-branch-name>"
	- Deploy a commit (8 characters or more), for example: "kyma deploy --source=34edf09a"
	- Deploy a pull request, for example "kyma deploy --source=PR-9486"
	- Deploy the local sources: "kyma deploy --source=local"`)

	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "", "Custom domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: %s, %s.", profileEvaluation, profileProduction))
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation.")
	cobraCmd.Flags().StringSliceVarP(&o.Values, "value", "", []string{}, "Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').")
	cobraCmd.Flags().StringSliceVarP(&o.ValueFiles, "values-file", "f", []string{}, "Path(s) to one or more JSON or YAML files with configuration values.")

	return cobraCmd
}

func (cmd *command) Run(o *Options) error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	l := logger.NewLogger(o.Verbose)
	ws, err := cmd.workspaceBuilder(l)
	if err != nil {
		return err
	}

	values, err := mergeValues(cmd.opts, ws, cmd.K8s.Static())
	if err != nil {
		return err
	}

	isK3d, err := k3d.IsK3dCluster(cmd.K8s.Static())
	if err != nil {
		return err
	}

	hasCustomDomain := cmd.opts.Domain != ""
	if _, err := coredns.Patch(zap.NewNop(), cmd.K8s.Static(), hasCustomDomain, isK3d); err != nil {
		return err
	}

	components, err := cmd.createCompListWithOverrides(ws, values)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace: %#v\n", ws)
	cmd.installPrerequisites(ws.WorkspaceDir)
	err = cmd.deployKyma(components)
	if err != nil {
		return err
	}

	if err := cmd.importCertificate(); err != nil {
		return err
	}

	// TODO: print summary after deploy

	return nil
}

func (cmd *command) deployKyma(comps component.List) error {
	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not read kubeconfig")
	}

	localScheduler := scheduler.NewLocalScheduler(
		scheduler.WithPrerequisites(cmd.buildCompList(comps.Prerequisites)...),
		scheduler.WithStatusFunc(cmd.printDeployStatus))

	componentsToInstall := append(comps.Prerequisites, comps.Components...)
	err = localScheduler.Run(context.TODO(), &keb.Cluster{
		Kubeconfig: string(kubeconfig),
		KymaConfig: keb.KymaConfig{
			Version:    cmd.opts.Source,
			Profile:    cmd.opts.Profile,
			Components: componentsToInstall,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to deploy Kyma")
	}
	return nil
}

func (cmd *command) workspaceBuilder(l *zap.SugaredLogger) (*workspace.Workspace, error) {
	wsStep := cmd.NewStep(fmt.Sprintf("Fetching Kyma (%s)", cmd.opts.Source))

	//Check if workspace is empty or not
	if cmd.opts.Source != VersionLocal {
		_, err := os.Stat(cmd.opts.WorkspacePath)
		// workspace already exists
		if !os.IsNotExist(err) && !cmd.avoidUserInteraction() {
			isWorkspaceEmpty, err := files.IsDirEmpty(cmd.opts.WorkspacePath)
			if err != nil {
				return nil, err
			}
			// if workspace used is not the default one and it is not empty,
			// then ask for permission to delete its existing files
			if !isWorkspaceEmpty && cmd.opts.WorkspacePath != getDefaultWorkspacePath() {
				if !wsStep.PromptYesNo(fmt.Sprintf("Existing files in workspace folder '%s' will be deleted. Are you sure you want to continue? ", cmd.opts.WorkspacePath)) {
					wsStep.Failure()
					return nil, fmt.Errorf("aborting deployment")
				}
			}
		}
	}

	wsp, err := cmd.opts.ResolveLocalWorkspacePath()
	if err != nil {
		return &workspace.Workspace{}, err
	}

	wsFact, err := workspace.NewFactory(nil, wsp, l)
	if err != nil {
		return &workspace.Workspace{}, err
	}
	err = service.UseGlobalWorkspaceFactory(wsFact)
	if err != nil {
		return nil, err
	}

	ws, err := wsFact.Get(cmd.opts.Source)
	if err != nil {
		return &workspace.Workspace{}, err
	}

	wsStep.Successf("Fetching kyma from workspace folder: %s", wsp)

	return ws, nil
}

func (cmd *command) buildCompList(comps []keb.Component) []string {
	var compSlice []string
	for _, c := range comps {
		compSlice = append(compSlice, c.Component)
	}
	return compSlice
}

func (cmd *command) createCompListWithOverrides(ws *workspace.Workspace, overrides map[string]interface{}) (component.List, error) {
	var compList component.List
	if len(cmd.opts.Components) > 0 {
		compList = component.FromStrings(cmd.opts.Components, overrides)
		return compList, nil
	}
	if cmd.opts.ComponentsFile != "" {
		return component.FromFile(ws, cmd.opts.ComponentsFile, overrides)
	}
	compFile := path.Join(ws.InstallationResourceDir, "components.yaml")
	return component.FromFile(ws, compFile, overrides)
}

func (cmd *command) printDeployStatus(component string, msg *reconciler.CallbackMessage) {
	fmt.Printf("Component %s has status %s\n", component, msg.Status)
}

// avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}

func (cmd *command) importCertificate() error {
	ca := trust.NewCertifier(cmd.K8s)

	if !cmd.approveImportCertificate() {
		//no approval given: stop import
		ca.InstructionsKyma2()
		return nil
	}

	// get cert from cluster
	cert, err := ca.CertificateKyma2()
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

	// create a simple step to print certificate import steps without a spinner (spinner overwrites sudo prompt)
	// TODO refactor how certifier logs when the old install command is gone
	f := step.Factory{
		NonInteractive: true,
	}
	s := f.NewStep("Importing Kyma certificate")

	if err := ca.StoreCertificate(tmpFile.Name(), s); err != nil {
		return err
	}
	s.Successf("Kyma root certificate imported")
	return nil
}

func (cmd *command) approveImportCertificate() bool {
	qImportCertsStep := cmd.NewStep("Install Kyma certificate locally")
	defer qImportCertsStep.Success()
	if cmd.avoidUserInteraction() { //do not import if user-interaction has to be avoided (suppress sudo pwd request)
		return false
	}
	return qImportCertsStep.PromptYesNo("Should the Kyma certificate be installed locally?")
}


type IstioConfig struct {
	APIVersion    string   `yaml:"apiVersion"`
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	AppVersion    string   `yaml:"appVersion"`
	TillerVersion string   `yaml:"tillerVersion"`
	Description   string   `yaml:"description"`
	Keywords      []string `yaml:"keywords"`
	Sources       []string `yaml:"sources"`
	Engine        string   `yaml:"engine"`
	Home          string   `yaml:"home"`
	Icon          string   `yaml:"icon"`
}

func (cmd *command) installPrerequisites(wsp string) {
	istioStep := cmd.NewStep("Installing Prerequisites")

	// Get wanted Istio Version
	var chart IstioConfig
	istioConfig, err := ioutil.ReadFile(filepath.Join(wsp, istioChartPath))
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		//TODO
	}
	err = yaml.Unmarshal(istioConfig, &chart)
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		//TODO
	}
	istioVersion := chart.AppVersion
	fmt.Printf("Istio Version: %v\n", istioVersion)
	// Get OS Version
	osext := runtime.GOOS
	switch osext {
	case "windows":
		osext = "win"
	case "darwin":
		osext = "osx"
	default:
		osext = "linux"
	}

	istioArch := runtime.GOARCH
	//if osext == "osx" && istioArch == "amd64"{
	//	istioArch = "arm64"
	//}

	nonArchUrl := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-%s.tar.gz", istioVersion, istioVersion, osext)
	archUrl := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-%s-%s.tar.gz", istioVersion, istioVersion, osext, istioArch)

	archSupport := "1.6"

	if osext == "linux" {
		if strings.Split(archSupport, ".")[1] >= strings.Split(istioVersion, ".")[1] {
			err := DownloadFile(path.Join(wsp, "istioctl"), "istio.tar.gz", archUrl)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
		} else {
			err := DownloadFile(path.Join(wsp, "istioctl"), "istio.tar.gz", nonArchUrl)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
		}
	} else if osext == "osx" {
		err := DownloadFile(path.Join(wsp, "istioctl"), "istio.tar.gz", nonArchUrl)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
	} else if osext == "win" {
		// TODO
	} else {
		// TODO
	}

	// Unzip tar.gz
	istioPath := path.Join(wsp, "istioctl", "istio.tar.gz")
	targetPath := path.Join(wsp, "istioctl", "istio.tar")
	fmt.Printf("IstioPath %s\n", istioPath)
	UnGzip(istioPath, targetPath)
	istioPath = path.Join(wsp, "istioctl", "istio.tar")
	targetPath = path.Join(wsp, "istioctl")
	Untar(istioPath, targetPath)

	// Export env variable
	binPath := path.Join(wsp, "istioctl", fmt.Sprintf("istio-%s", istioVersion), "bin", "istioctl")
	os.Setenv("ISTIOCTL_PATH", binPath)
	fmt.Printf("Env: %s\n", os.Getenv("ISTIOCTL_PATH"))
	istioStep.Success()
}


func DownloadFile(filepath string, filename string, url string) error {
	// Get data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create path and file
	os.MkdirAll(filepath, 0700)
	out, err := os.Create(path.Join(filepath, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	// Write body to file
	_, err = io.Copy(out, resp.Body)
	return err
}


func UnGzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
