package uninstall

import (
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/spf13/cobra"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/overrides"
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

	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 1200*time.Second, "Maximum time for the deletion")
	cobraCmd.Flags().DurationVarP(&o.TimeoutComponent, "timeout-component", "", 360*time.Second, "Maximum time to delete the component")
	cobraCmd.Flags().IntVar(&o.Concurrency, "concurrency", 4, "Number of parallel processes")
	cobraCmd.Flags().BoolVarP(&o.KeepCRDs, "keep-crds", "", false, "Flag specifying whether to keep CRDs on deletion")
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
		KeepCRDs:                      cmd.opts.KeepCRDs,
		KubeconfigSource: installConfig.KubeconfigSource{
			Path: kube.KubeconfigPath(cmd.KubeconfigPath),
		},
	}

	// if not verbose, use asyncui for clean output
	var callback func(deployment.ProcessUpdate)
	if !cmd.Verbose {
		ui := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		callback = ui.Callback()
		if err != nil {
			return err
		}
	}

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(installCfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(installCfg.BackoffMaxElapsedTimeSeconds / installCfg.BackoffInitialIntervalSeconds)),
		retry.DelayType(retry.FixedDelay),
	}

	installer, err := deployment.NewDeletion(installCfg, &overrides.Builder{}, callback, commonRetryOpts)
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
	metaProv, err := helm.NewKymaMetadataProvider(installConfig.KubeconfigSource{
		Path: kube.KubeconfigPath(cmd.KubeconfigPath),
	})
	if err != nil {
		return nil, err
	}

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
