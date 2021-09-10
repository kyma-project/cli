package deploy

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/components"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/k3d"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/overrides"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/fs"
	"io/ioutil"
	"os"
	"path"

	//Register all reconcilers
	_ "github.com/kyma-incubator/reconciler/pkg/reconciler/instances"
)

const defaultVersion = "main"
const defaultProfile = "evaluation"

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
	return cobraCmd
}



func (cmd *command) createComplistWithOverrides(ws *workspace.Workspace,  overrides map[string]interface{}) (components.ComponentList, error) {
	var compList components.ComponentList
	if len(cmd.opts.Components) > 0 {
		compList = components.ComponentsFromStrings(cmd.opts.Components, overrides)
		return compList , nil
	}
	compFile := cmd.opts.ResolveComponentsFile(ws)
	compList, err := components.NewComponentList(compFile, overrides)
	if err != nil {
			return compList, err
	}
	return  compList, nil
}

func (cmd *command) Run(o *Options) error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	ws, err := cmd.loadWorkspace()
	if err != nil {
		return err
	}

	ovs, err := cmd.buildOverrides(ws)
	if err != nil {
		return err
	}

	isK3d, err := k3d.IsK3dCluster(cmd.K8s.Static())
	if err != nil {
		return err
	}

	if _, err := coredns.Patch(zap.NewNop(), cmd.K8s.Static(), ovs, isK3d); err != nil {
		return err
	}

	comps, err := cmd.createComplistWithOverrides(ws, ovs.FlattenedMap())
	if err != nil {
		return err
	}

	err = cmd.deployKyma(comps)
	if err != nil {
		return err
	}

	if err := cmd.importCertificate(); err != nil {
		return err
	}

	// TODO: print summary after deploy

	return nil
}

func (cmd *command) loadWorkspace() (*workspace.Workspace, error) {
	downloadStep := cmd.NewStep(fmt.Sprintf("Downloading Kyma (%s) into workspace folder ", defaultVersion))

	workspaceDir, err := files.KymaHome()
	if err != nil {
		return nil, errors.Wrap(err, "Could not create Kyma home directory")
	}

	//use a global workspace factory to ensure all component-reconcilers are using the same workspace-directory
	//(otherwise each component-reconciler would handle the download of Kyma resources individually which will cause
	//collisions when sharing the same directory)
	factory, err := workspace.NewFactory(workspaceDir, zap.NewNop().Sugar())
	if err != nil {
		return nil, err
	}

	err = service.UseGlobalWorkspaceFactory(factory)
	if err != nil {
		return nil, err
	}

	ws, err := factory.Get(defaultVersion)
	if err != nil {
		return nil, err
	}

	downloadStep.Successf("Kyma downloaded into workspace folder")

	return ws, nil
}

func (cmd *command) buildOverrides(workspace *workspace.Workspace) (overrides.Overrides, error) {
	overridesStep := cmd.NewStep("Building Kyma2 overrides")

	overridesBuilder := &overrides.Builder{}

	kyma2OverridesPath := path.Join(workspace.InstallationResourceDir, "values.yaml")

	if err := overridesBuilder.AddFile(kyma2OverridesPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			overridesStep.LogInfof("Kyma2 override path not found but continuing: %s", err)
		} else {
			return overrides.Overrides{}, errors.Wrap(err, "Could not add overrides for Kyma 2.0")
		}
	}

	overridesBuilder.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, overrides.NewDomainNameOverrideInterceptor(cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, overrides.NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, overrides.NewRegistryInterceptor(cmd.K8s.Static()))
	overridesBuilder.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, overrides.NewRegistryDisableInterceptor(cmd.K8s.Static()))

	ovs, err := overridesBuilder.Build()
	if err != nil {
		return overrides.Overrides{}, err
	}

	overridesStep.Success()
	return ovs, err
}

func (cmd *command) deployKyma(comps components.ComponentList) error {
	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not read kubeconfig")
	}
	fmt.Printf("PRREQ: %v \n", components.BuildCompList(comps.Prerequisites))
	localScheduler := scheduler.NewLocalScheduler(
		scheduler.WithCRDComponents("cluster-essentials"),
		scheduler.WithPrerequisites(components.BuildCompList(comps.Prerequisites)...),
		scheduler.WithStatusFunc(cmd.printDeployStatus))

	componentsToInstall := append(comps.Prerequisites, comps.Components...)
	fmt.Printf("COMPS TO INSTALL: %v \n", componentsToInstall)
	err = localScheduler.Run(context.TODO(), &keb.Cluster{
		Kubeconfig: string(kubeconfig),
		KymaConfig: keb.KymaConfig{
			Version:    defaultVersion,
			Profile:    defaultProfile,
			Components: componentsToInstall,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to deploy Kyma")
	}
	return nil
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
