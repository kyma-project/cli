package uninstall

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/spf13/cobra"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
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
		Use:     "delete",
		Short:   "Deletes Kyma from a running Kubernetes cluster.",
		Long:    `Use this command to delete Kyma from a running Kubernetes cluster.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"d"},
	}

	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", defaultWorkspacePath, "Path used to download Kyma sources.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 1200*time.Second, "Maximum time for the deletion (default: 20m0s)")
	cobraCmd.Flags().DurationVarP(&o.TimeoutComponent, "timeout-component", "", 360*time.Second, "Maximum time to delete the component (default: 6m0s)")
	cobraCmd.Flags().IntVar(&o.Concurrency, "concurrency", 4, "Number of parallel processes (default: 4)")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	var ui asyncui.AsyncUI
	if !cmd.Verbose { //use async UI only if not in verbose mode
		ui = asyncui.AsyncUI{StepFactory: &cmd.Factory}
		if err := ui.Start(); err != nil {
			return err
		}
		defer ui.Stop()
	}

	//get list of installed Kyma components
	compList, err := cmd.kymaComponentList()
	if err != nil {
		return err
	}

	installCfg := &installConfig.Config{
		WorkersCount:                  cmd.opts.Concurrency,
		CancelTimeout:                 cmd.opts.Timeout,
		QuitTimeout:                   cmd.opts.QuitTimeout(),
		HelmTimeoutSeconds:            int(cmd.opts.TimeoutComponent.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.NewHydroformLoggerAdapter(cli.NewLogger(cmd.Verbose)),
		ComponentList:                 compList,
	}

	// if an AsyncUI is used, get channel for update events
	var updateCh chan<- deployment.ProcessUpdate
	if ui.IsRunning() {
		updateCh, err = ui.UpdateChannel()
		if err != nil {
			return err
		}
	}

	installer, err := deployment.NewDeletion(installCfg, &deployment.OverridesBuilder{}, cmd.K8s.Static(), updateCh)
	if err != nil {
		return err
	}

	uninstallErr := installer.StartKymaUninstallation()

	if uninstallErr == nil {
		cmd.showSuccessMessage()
	}
	return uninstallErr
}

func (cmd *command) kymaComponentList() (*installConfig.ComponentList, error) {
	kymaCompStep := cmd.NewStep("Get Kyma components")
	metaProv := helm.NewKymaMetadataProvider(cmd.K8s.Static())

	versionSet, err := metaProv.Versions()
	if err != nil {
		kymaCompStep.Failure()
		return nil, err
	}

	compList := &installConfig.ComponentList{}
	for _, comp := range versionSet.InstalledComponents() {
		compDef := installConfig.ComponentDefinition{
			Name:      comp.Name,
			Namespace: comp.Namespace,
		}
		if comp.Prerequisite {
			compList.Prerequisites = append(compList.Prerequisites, compDef)
		} else {
			compList.Components = append(compList.Components, compDef)
		}
	}

	kymaCompStep.Successf("Found %d Kyma versions (%s) and %d Kyma components",
		versionSet.Count(), strings.Join(versionSet.Names(), ", "), len(compList.Components))

	return compList, nil
}

func (cmd *command) showSuccessMessage() {
	// TODO: show processing summary
	fmt.Println("Kyma successfully removed.")
}
