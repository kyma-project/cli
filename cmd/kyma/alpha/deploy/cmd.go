package deploy

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"errors"

	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/pkg/dashboard"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
)

type command struct {
	cli.Command
	opts *Options
}

const (
	lifecycleManagerKustomization = "https://github.com/kyma-project/lifecycle-manager/config/default"

	hostsTemplate = `
    {{ .K3dRegistryIP}} {{ .K3dRegistryHost}}
    {{ .K3dRegistryIP}} {{ .K3dRegistryHost}}.localhost
`
)

// NewCmd creates a new deploy command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(cobraCmd *cobra.Command, _ []string) error { return cmd.RunWithTimeout(cobraCmd.Context()) },
		Aliases: []string{"d"},
		Example: `
- Deploy the latest version of the Lifecycle Manager for trying out Modules: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default -with-wildcard-permissions"
- Deploy the main branch of Lifecycle Manager: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default@main"
- Deploy a local version of Lifecycle Manager: "kyma deploy -k /path/to/repo/lifecycle-manager/config/default"`,
	}
	cobraCmd.Flags().StringArrayVarP(
		&o.Kustomizations, "kustomization", "k", []string{lifecycleManagerKustomization},
		`Provide one or more kustomizations to deploy. 
Each occurrence of the flag accepts a URL with an optional reference (commit, branch, or release) in the format URL@ref or a local path to the directory of the kustomization file.
Defaults to deploying Lifecycle Manager and Module Manager from GitHub main branch.`,
	)
	cobraCmd.Flags().StringArrayVarP(
		&o.Modules, "module", "m", []string{},
		`Provide one or more modules to activate after the deployment is finished. Example: "--module name@namespace" (namespace is optional).`,
	)
	cobraCmd.Flags().StringVarP(&o.ModulesFile, "modules-file", "f", "", `Path to file containing a list of modules.`)
	cobraCmd.Flags().StringVarP(
		&o.Channel, "channel", "c", "regular", `Select which channel to deploy from.`,
	)
	cobraCmd.Flags().StringVarP(
		&o.Namespace, "namespace", "n", cli.KymaNamespaceDefault,
		"The Namespace to deploy the the Kyma custom resource in.",
	)
	cobraCmd.Flags().StringVar(&o.KymaCR, "kyma-cr", "", `Provide a custom Kyma CR file for the deployment.`)

	// TODO remove this flag when module templates can be fetched from release.
	// Might be worth keeping this flag with another name to install extra templates??
	cobraCmd.Flags().StringArrayVar(
		&o.Templates, "template", []string{}, `Provide one or more module templates to deploy.
WARNING: This is a temporary flag for development and will be removed soon.`,
	)

	cobraCmd.Flags().StringVar(
		&o.CertManagerVersion, "cert-manager", "v1.11.0",
		"Installs cert-manager from the specified static version. an empty string skips the installation.",
	)
	cobraCmd.Flags().StringVar(
		&o.LifecycleManager, "lifecycle-manager", "eu.gcr.io/kyma-project/lifecycle-manager:latest",
		`Installs lifecycle-manager with the specified image:
- Use "my-registry.org/lifecycle-manager:my-tag"" to use a custom version of lifecycle-manager.
- Use "europe-docker.pkg.dev/kyma-project/prod/lifecycle-manager@sha256:cb74b29cfe80c639c9ee9..." to use a custom version of lifecycle-manager with a digest.
- Use a tag like "v20230220-7b8e9515" to override the default tag used, it will then use "eu.gcr.io/kyma-project/lifecycle-manager:v20230220-7b8e9515".`,
	)

	cobraCmd.Flags().BoolVar(
		&o.DryRun, "dry-run", false, "Renders the Kubernetes manifests without actually applying them.",
	)

	cobraCmd.Flags().BoolVar(
		&o.WildcardPermissions, "wildcard-permissions", true,
		`Creates a wildcard cluster-role to allow for easy local installation permissions of lifecycle-manager.
Allows for usage of lifecycle-manager without having to worry about modules requiring specific RBAC permissions.
WARNING: DO NOT USE ON PRODUCTIVE CLUSTERS!`,
	)

	cobraCmd.Flags().BoolVar(
		&o.OpenDashboard, "open-dashboard", false,
		`Opens the Busola Dashboard at startup. Only works when a graphical interface is available and when running in interactive mode`,
	)

	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "t", 20*time.Minute, "Maximum time for the deployment.")

	return cobraCmd
}

func (cmd *command) RunWithTimeout(ctx context.Context) error {
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

	err := cmd.run(ctx)

	// yes, I tried errors.As and errors.Is, and both did not work or threw vet issues...
	if err != nil && strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		msg := "Timeout reached while waiting for deployment to complete"
		timeoutStep := cmd.NewStep(msg)
		timeoutStep.Failure()
		return fmt.Errorf("%s: %w", msg, err)
	}

	return err
}

func (cmd *command) run(ctx context.Context) error {
	start := time.Now()

	var err error
	if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
		return fmt.Errorf("failed to initialize the Kubernetes client from given kubeconfig: %w", err)
	}

	if err := cmd.deploy(ctx, start); err != nil {
		return err
	}

	// do not starrt the dashboard if not interactive
	if cmd.opts.CI || cmd.opts.NonInteractive || !cmd.opts.OpenDashboard {
		return nil
	}

	return cmd.wizard(ctx)
}

func (cmd *command) deploy(ctx context.Context, start time.Time) error {
	if cmd.opts.DryRun {
		return cmd.dryRun(ctx)
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()

	summary := &nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        "alpha deployment with lifecycle-manager",
	}

	undo := zap.RedirectStdLog(l.Desugar())
	defer undo()

	if !cmd.opts.Verbose {
		stderr := os.Stderr
		os.Stderr = nil
		defer func() { os.Stderr = stderr }()
	}

	clusterInfo, err := clusterinfo.Discover(ctx, cmd.K8s.Static())
	if err != nil {
		return err
	}

	if cmd.opts.CertManagerVersion != "" {
		certManagerStep := cmd.NewStep("Deploying cert-manager.io")
		certManagerStep.Start()
		if err := deploy.CertManager(ctx, cmd.K8s, cmd.opts.CertManagerVersion, false); err != nil {
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

	deployStep := cmd.NewStep("Deploying Kustomizations")
	deployStep.Start()
	hasKyma, err := deploy.Bootstrap(
		ctx, cmd.opts.Kustomizations, cmd.K8s, cmd.opts.Filters, cmd.opts.WildcardPermissions, false,
	)
	if err != nil {
		deployStep.Failuref("Failed to deploy Kustomizations %s: %s", cmd.opts.Kustomizations, err.Error())
		return err
	}
	deployStep.Successf("Kustomizations deployed: %s", cmd.opts.Kustomizations)

	coreDNS := cmd.NewStep("Patching CoreDNS")
	coreDNS.Start()
	if _, err := coredns.Patch(l.Desugar(), cmd.K8s.Static(), false, clusterInfo, hostsTemplate); err != nil {
		coreDNS.Failuref("error patching CoreDNS: %s", err)
		return err
	}
	coreDNS.Successf("CoreDNS patched successfully")

	// deploy modules and kyma CR
	if hasKyma {
		// TODO change to fetch templates from release artifacts
		if len(cmd.opts.Templates) > 0 {
			modStep := cmd.NewStep("Module Templates deployed")
			for _, t := range cmd.opts.Templates {
				b, err := os.ReadFile(t)
				if err != nil {
					modStep.Failuref("Failed to deploy module templates")
					return err
				}
				resources, err := cmd.K8s.ParseManifest(b)
				if err != nil {
					modStep.Failuref("Failed to parse manifest for module templates")
					return err
				}
				if err := cmd.K8s.Apply(ctx, resources); err != nil {
					modStep.Failuref("Failed to deploy module templates")
					return err
				}
			}
			modStep.Success()
		}

		kymaStep := cmd.NewStep("Deploying Kyma CR")
		kymaStep.Start()
		if err := deploy.Kyma(ctx, cmd.K8s, cmd.opts.Namespace, cmd.opts.Channel, cmd.opts.KymaCR, false); err != nil {
			kymaStep.Failuref("Failed to deploy Kyma CR: %s", err.Error())
			return err
		}
		kymaStep.Successf("Kyma CR deployed and Ready!")
	}

	deployTime := time.Since(start)
	return summary.Print(deployTime)
}

func (cmd *command) dryRun(ctx context.Context) error {
	if cmd.opts.CertManagerVersion != "" {
		if err := deploy.CertManager(ctx, cmd.K8s, cmd.opts.CertManagerVersion, true); err != nil {
			return err
		}
	}

	hasKyma, err := deploy.Bootstrap(
		ctx, cmd.opts.Kustomizations, cmd.K8s, cmd.opts.Filters, cmd.opts.WildcardPermissions, true,
	)
	if err != nil {
		return err
	}

	if hasKyma {
		// TODO change to fetch templates from release artifacts
		for _, t := range cmd.opts.Templates {
			b, err := os.ReadFile(t)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n---\n", string(b))
		}

		if err := deploy.Kyma(
			ctx, cmd.K8s, cmd.opts.Namespace, cmd.opts.Channel, cmd.opts.KymaCR, true,
		); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) wizard(ctx context.Context) error {
	// get all infos for the dashboard URL
	ctx, cancel := context.WithTimeout(ctx, cmd.opts.Timeout)
	defer cancel()

	kymas, err := cmd.K8s.Dynamic().Resource(
		schema.GroupVersionResource{
			Group:    "operator.kyma-project.io",
			Version:  "v1alpha1",
			Resource: "kymas",
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
	// make sure the dahboard container always stops at the end and the cursor restored
	cmd.Finalizers.Add(dash.StopFunc(ctx, func(i ...interface{}) { fmt.Print(i...) }))

	if err := dash.Open(fmt.Sprintf("/cluster/%s/namespaces/%s/kymas/details/%s", cluster, ns, name)); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Successf("Dashboard started. To exit press Ctrl+C")

	return dash.Watch(ctx)
}
