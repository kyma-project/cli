package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"golang.org/x/exp/slices"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/pkg/dashboard"
)

type command struct {
	cli.Command
	opts *Options
}

const (
	lifecycleManagerKustomization = "https://github.com/kyma-project/lifecycle-manager/config/default"
	hostsTemplate                 = `
    {{ .K3dRegistryIP}} {{ .K3dRegistryHost}}
    {{ .K3dRegistryIP}} {{ .K3dRegistryHost}}.localhost
`
)

func NewCmd(o *Options) *cobra.Command {
	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}
	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(cobraCmd *cobra.Command, _ []string) error { return cmd.runWithTimeout(cobraCmd.Context()) },
		Aliases: []string{"d"},
		Example: `
- Deploy the latest version of the Lifecycle Manager for trying out Modules: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default -with-wildcard-permissions"
- Deploy the main branch of Lifecycle Manager: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default@main"
- Deploy a local version of Lifecycle Manager: "kyma deploy -k /path/to/repo/lifecycle-manager/config/default"`,
	}
	cobraCmd.Flags().StringArrayVarP(
		&o.Kustomizations,
		"kustomization",
		"k",
		[]string{lifecycleManagerKustomization},
		`Provide one or more kustomizations to deploy. 
Each flag occurrence accepts a URL with an optional reference (commit, branch, or release) in URL@ref format or a local path to the directory of the kustomization file.
By default, Lifecycle Manager is deployed from the GitHub main branch.`,
	)
	cobraCmd.Flags().StringArrayVarP(
		&o.Modules,
		"module",
		"m",
		[]string{},
		`Provide one or more modules to activate after the deployment is finished. Example: "--module name@namespace" (namespace is optional).`,
	)
	cobraCmd.Flags().StringVarP(
		&o.Channel,
		"channel",
		"c",
		"regular",
		`Selects which channel to deploy from.`,
	)
	cobraCmd.Flags().StringVarP(
		&o.Namespace,
		"namespace",
		"n",
		cli.KymaNamespaceDefault,
		"The Namespace to deploy the Kyma custom resource in.",
	)
	cobraCmd.Flags().StringVar(
		&o.KymaCR,
		"kyma-cr",
		"",
		`Provide a custom Kyma CR file for the deployment.`,
	)
	cobraCmd.Flags().StringArrayVar(
		&o.AdditionalTemplates,
		"extra-templates",
		[]string{},
		`Provide one or more additional module templates via URL or local path to apply after deployment.`,
	)
	cobraCmd.Flags().StringVar(
		&o.CertManagerVersion,
		"cert-manager",
		"v1.11.0",
		"Installs cert-manager from the specified static version. An empty string skips the installation.",
	)
	cobraCmd.Flags().StringVar(
		&o.LifecycleManager,
		"lifecycle-manager",
		"",
		`Installs Lifecycle Manager with the specified image:
- Use "my-registry.org/lifecycle-manager:my-tag"" to use a custom version of Lifecycle Manager.
- Use "europe-docker.pkg.dev/kyma-project/prod/lifecycle-manager@sha256:cb74b29cfe80c639c9ee9..." to use a custom version of Lifecycle Manager with a digest.
- Specify a tag to override the default one. For example, when specifying "v20230220-7b8e9515", the "eu.gcr.io/kyma-project/lifecycle-manager:v20230220-7b8e9515" tag is used.`,
	)
	cobraCmd.Flags().BoolVar(
		&o.DryRun,
		"dry-run",
		false,
		"Renders the Kubernetes manifests without actually applying them.",
	)
	cobraCmd.Flags().BoolVar(
		&o.WildcardPermissions,
		"wildcard-permissions",
		true,
		`Creates a wildcard cluster-role to allow for easy local installation permissions of Lifecycle Manager.
Allows for Lifecycle Manager usage without worrying about modules requiring specific RBAC permissions.
WARNING: DO NOT USE ON PRODUCTIVE CLUSTERS!`,
	)
	cobraCmd.Flags().BoolVar(
		&o.OpenDashboard,
		"open-dashboard",
		false,
		`Opens the Busola Dashboard at startup. Only works when a graphical interface is available and when running in interactive mode`,
	)
	cobraCmd.Flags().BoolVarP(
		&o.Force,
		"force-conflicts",
		"f",
		false,
		"Forces the patching of Kyma spec modules in case their managed field was edited by a source other than Kyma CLI.",
	)
	cobraCmd.Flags().StringVar(
		&o.Target,
		"target",
		targetControlPlane,
		"Target to use when determining where to install default modules. Available values are 'control-plane' or 'remote'.",
	)
	cobraCmd.Flags().DurationVarP(
		&o.Timeout,
		"timeout",
		"t", 20*time.Minute,
		"Maximum time for the deployment.",
	)

	return cobraCmd
}

func (cmd *command) runWithTimeout(ctx context.Context) error {
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if err := cmd.opts.validateFlags(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, cmd.opts.Timeout)
	defer cancel()

	if err := cmd.deploy(ctx); err != nil {
		return cmd.handleTimeoutErr(err)
	}

	// skip dashboard if non-interactive
	if cmd.opts.CI || cmd.opts.NonInteractive || !cmd.opts.OpenDashboard {
		return nil
	}

	if err := cmd.openDashboard(ctx); err != nil {
		return cmd.handleTimeoutErr(err)
	}

	return nil
}

func (cmd *command) deploy(ctx context.Context) error {
	isInKcpMode := cmd.opts.Target == targetRemote
	if cmd.opts.DryRun {
		return cmd.dryRun(ctx, isInKcpMode)
	}

	start := time.Now()
	log := cli.NewLogger(cmd.opts.Verbose).Sugar()
	undo := zap.RedirectStdLog(log.Desugar())
	defer undo()

	clusterAccess := cmd.NewStep("Ensuring Cluster Access")
	info, err := cmd.EnsureClusterAccess(ctx, cmd.opts.Timeout)
	if err != nil {
		clusterAccess.Failuref("Could not ensure cluster Access")
		return err
	}
	clusterAccess.Successf("Successfully connected to cluster")

	if !cmd.opts.CI && !cmd.opts.NonInteractive {
		if err := cmd.detectManagedKyma(ctx); err != nil {
			return err
		}
	}

	if !cmd.opts.Verbose {
		stderr := os.Stderr
		os.Stderr = nil
		defer func() { os.Stderr = stderr }()
	}

	if cmd.opts.CertManagerVersion != "" {
		cmd.deployCertManager(ctx)
	}

	deployStep := cmd.NewStep("Deploying Kustomizations")
	deployStep.Start()
	hasKyma, err := deploy.Bootstrap(
		ctx, cmd.opts.Kustomizations, cmd.K8s, cmd.opts.Filters, cmd.opts.WildcardPermissions, cmd.opts.Force, false,
		isInKcpMode,
	)
	if err != nil {
		deployStep.Failuref("Failed to deploy Kustomizations %s: %s", cmd.opts.Kustomizations, err.Error())
		return err
	}
	deployStep.Successf("Kustomizations deployed: %s", cmd.opts.Kustomizations)

	coreDNS := cmd.NewStep("Patching CoreDNS")
	coreDNS.Start()
	if _, err := coredns.Patch(log.Desugar(), cmd.K8s.Static(), false, info, hostsTemplate); err != nil {
		coreDNS.Failuref("error patching CoreDNS: %s", err)
		return err
	}
	coreDNS.Successf("CoreDNS patched successfully")

	if !hasKyma {
		// skip applying Kyma CR and module templates
		return cmd.printSummary(start)
	}

	if len(cmd.opts.AdditionalTemplates) > 0 {
		modStep := cmd.NewStep("Deploying additional module templates")
		if err := deploy.ModuleTemplates(ctx, cmd.K8s, cmd.opts.AdditionalTemplates, cmd.opts.Target, cmd.opts.Force,
			false); err != nil {
			modStep.Failuref("Failed to deploy additional module templates")
			return err
		}
		modStep.Successf("Additional module templates deployed: %s", cmd.opts.AdditionalTemplates)
	}

	kymaStep := cmd.NewStep("Deploying Kyma CR")
	if err := deploy.Kyma(
		ctx, cmd.K8s, cmd.opts.Namespace, cmd.opts.Channel, cmd.opts.KymaCR, cmd.opts.Force, false, isInKcpMode,
	); err != nil {
		kymaStep.Failuref("Failed to deploy Kyma CR: %s", err.Error())
		return err
	}
	kymaStep.Successf("Kyma CR deployed and Ready!")

	return cmd.printSummary(start)
}

func (cmd *command) printSummary(startTime time.Time) error {
	deployTime := time.Since(startTime)
	summary := &nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        "alpha deployment with lifecycle-manager",
	}
	return summary.Print(deployTime)
}

func (cmd *command) deployCertManager(ctx context.Context) {
	certManagerStep := cmd.NewStep("Deploying cert-manager.io")
	certManagerStep.Start()
	if err := deploy.CertManager(ctx, cmd.K8s, cmd.opts.CertManagerVersion, cmd.opts.Force, false); err != nil {
		certManagerStep.LogWarn(err.Error())
		certManagerStep.Failuref("Failed to deploy cert-manager.io.")
	}
	err := cmd.K8s.WaitDeploymentStatus(
		"cert-manager", "cert-manager-webhook", appsv1.DeploymentAvailable, corev1.ConditionTrue,
	)
	if err != nil {
		certManagerStep.LogWarn(err.Error())
		certManagerStep.Failuref("cert-manager.io webhook failed to start.")
	}
	certManagerStep.Successf(
		"Deployed cert-manager.io in version %s",
		cmd.opts.CertManagerVersion,
	)
}

func (cmd *command) dryRun(ctx context.Context, isInKcpMode bool) error {
	if cmd.opts.CertManagerVersion != "" {
		if err := deploy.CertManager(ctx, cmd.K8s, cmd.opts.CertManagerVersion, cmd.opts.Force, true); err != nil {
			return err
		}
	}

	hasKyma, err := deploy.Bootstrap(
		ctx, cmd.opts.Kustomizations, cmd.K8s, cmd.opts.Filters, cmd.opts.WildcardPermissions, cmd.opts.Force, true,
		isInKcpMode,
	)
	if err != nil {
		return err
	}

	if !hasKyma {
		return nil
	}

	if len(cmd.opts.AdditionalTemplates) > 0 {
		if err := deploy.ModuleTemplates(ctx, cmd.K8s, cmd.opts.AdditionalTemplates, cmd.opts.Target, cmd.opts.Force,
			true); err != nil {
			return err
		}
	}

	return deploy.Kyma(
		ctx, cmd.K8s, cmd.opts.Namespace, cmd.opts.Channel, cmd.opts.KymaCR, cmd.opts.Force, true, isInKcpMode,
	)
}

func (cmd *command) openDashboard(ctx context.Context) error {
	// get all infos for the dashboard URL
	ctx, cancel := context.WithTimeout(ctx, cmd.opts.Timeout)
	defer cancel()

	kymas, err := cmd.K8s.Dynamic().Resource(
		schema.GroupVersionResource{
			Group:    shared.OperatorGroup,
			Version:  v1beta2.GroupVersion.Version,
			Resource: shared.KymaKind.Plural(),
		},
	).List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	if len(kymas.Items) < 1 {
		return errors.New("no Kyma CR found in cluster")
	}

	cluster := cmd.K8s.KubeConfig().CurrentContext
	name := kymas.Items[0].GetName()
	ns := kymas.Items[0].GetNamespace()

	cmd.NewStep("Start module dashboard ")
	dash := dashboard.New("kyma-dashboard", "3001", cmd.KubeconfigPath, cmd.Verbose)
	if err := dash.Start(); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	// make sure the dashboard container always stops at the end and the cursor restored
	cmd.Finalizers.Add(dash.StopFunc(ctx, func(i ...interface{}) { fmt.Print(i...) }))

	if err := dash.Open(fmt.Sprintf("/cluster/%s/namespaces/%s/kymas/details/%s", cluster, ns, name)); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Successf("Dashboard started. To exit press Ctrl+C")

	return dash.Watch(ctx)
}

func (cmd *command) handleTimeoutErr(err error) error {
	if err == nil {
		return nil
	}

	// errors.As and errors.Is both do not work or throw vet issues
	if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		msg := "Timeout reached while waiting for deployment to complete"
		timeoutStep := cmd.NewStep(msg)
		timeoutStep.Failure()
		return fmt.Errorf("%s: %w", msg, err)
	}

	return err
}

func (cmd *command) detectManagedKyma(ctx context.Context) error {
	kymaResource := schema.GroupVersionResource{
		Group:    shared.OperatorGroup,
		Version:  v1beta2.GroupVersion.Version,
		Resource: shared.KymaKind.Plural(),
	}

	deployedKymaCRs, err := cmd.K8s.Dynamic().Resource(kymaResource).List(ctx, v1.ListOptions{})
	if err != nil || deployedKymaCRs == nil || len(deployedKymaCRs.Items) == 0 {
		return nil
	}

	kymaManagedFields := deployedKymaCRs.Items[0].GetManagedFields()
	if len(kymaManagedFields) == 0 {
		return nil
	}

	managedKymaField := v1.ManagedFieldsEntry{
		Manager:     "lifecycle-manager",
		Subresource: "status",
	}
	unmanagedKymaField := v1.ManagedFieldsEntry{
		Manager:     "unmanaged-kyma",
		Subresource: "status",
	}

	managedKymaWarning := "CAUTION: You are trying to use Kyma CLI to change a managed Kyma Resource. This action may corrupt the Kyma runtime. Proceed at your own risk."

	if slices.ContainsFunc(kymaManagedFields, func(field v1.ManagedFieldsEntry) bool {
		return field.Subresource == managedKymaField.Subresource && field.Manager == managedKymaField.Manager
	}) && !slices.ContainsFunc(kymaManagedFields, func(field v1.ManagedFieldsEntry) bool {
		return field.Subresource == unmanagedKymaField.Subresource && field.Manager == unmanagedKymaField.Manager
	}) {
		cmd.CurrentStep.LogInfo(managedKymaWarning)
		if !cmd.CurrentStep.PromptYesNo("Do you really want to proceed? ") {
			return errors.New("command stopped by user")
		}
	}

	return nil
}
