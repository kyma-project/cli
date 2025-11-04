package authorize

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type authorizeConfig struct {
	*cmdcommon.KymaConfig

	authTarget   string // user, group, serviceaccount
	name         []string
	namespace    string
	clusterWide  bool
	role         string
	clusterrole  string
	dryRun       bool
	outputFormat types.Format
}

func NewAuthorizeCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := authorizeConfig{
		KymaConfig: kymaConfig,
	}

	validAuthTargets := []string{"user", "group", "serviceaccount"}

	cmd := &cobra.Command{
		Use:   "authorize <authTarget> [flags]",
		Short: "Authorize a subject (user, group, or service account) with Kyma RBAC resources",
		Long:  "Create RoleBinding or ClusterRoleBinding granting access to a Kyma role / cluster role for a user, group, or service account.",
		Example: `  # Bind a user to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice --role view --namespace dev

  # Bind multiple users to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice,bob,james --role view --namespace dev

  # Bind a group cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize group --name team-observability --clusterrole kyma-read-all --cluster-wide

  # Bind a service account to a ClusterRole within a namespace (RoleBinding referencing a ClusterRole)
  kyma alpha authorize serviceaccount --name deployer-sa --clusterrole edit --namespace staging

  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize user --name bob --role operator --namespace ops --dry-run -o yaml

  # Generate JSON for a cluster-wide binding
  kyma alpha authorize user --name ci-bot --clusterrole kyma-admin --cluster-wide -o json`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: validAuthTargets,
		PreRun: func(cmd *cobra.Command, args []string) {
			cfg.authTarget = args[0]
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("name"),
				flags.MarkExactlyOneRequired("role", "clusterrole"),
				flags.MarkPrerequisites("role", "namespace"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorize(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace for RoleBinding (required unless --cluster-wide)")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "If true, create a ClusterRoleBinding; otherwise, a RoleBinding")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (namespaced)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (usable for RoleBinding or ClusterRoleBinding)")
	cmd.Flags().StringSliceVar(&cfg.name, "name", []string{}, "Name of the authorized subject(s)")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Print resources without applying")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (yaml or json)")

	cmd.AddCommand(NewAuthorizeRepositoryCMD(kymaConfig))

	return cmd
}

func authorize(cfg *authorizeConfig) clierror.Error {
	rbacResources, err := buildRBACResources(cfg)
	if err != nil {
		return err
	}

	if cfg.dryRun {
		return outputResources(cfg.outputFormat, rbacResources)
	}

	for _, res := range rbacResources {
		if applyErr := applyRBAC(cfg, res); applyErr != nil {
			return applyErr
		}
	}

	return nil
}

func buildRBACResources(cfg *authorizeConfig) ([]*unstructured.Unstructured, clierror.Error) {
	var rbacResources []*unstructured.Unstructured

	for _, subjectName := range cfg.name {
		subjectKind, err := authorization.NewSubjectKindFrom(cfg.authTarget)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("invalid authorization target"))
		}

		builder := authorization.NewRBACBuilder().
			ForSubjectKind(subjectKind).
			ForSubjectName(subjectName)

		var (
			rbac     *unstructured.Unstructured
			errBuild clierror.Error
		)

		if cfg.clusterWide {
			rbac, errBuild = buildClusterRoleBinding(builder, cfg, subjectName)
		} else {
			rbac, errBuild = buildRoleBinding(builder, cfg, subjectName)
		}

		if errBuild != nil {
			return nil, errBuild
		}

		rbacResources = append(rbacResources, rbac)
	}

	return rbacResources, nil
}

func buildClusterRoleBinding(builder *authorization.RBACBuilder, cfg *authorizeConfig, subjectName string) (*unstructured.Unstructured, clierror.Error) {
	rbacResource, err := builder.
		ForClusterRole(cfg.clusterrole).
		ForBindingName(fmt.Sprintf("%s-%s-binding", cfg.clusterrole, subjectName)).
		BuildClusterRoleBinding()

	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func buildRoleBinding(builder *authorization.RBACBuilder, cfg *authorizeConfig, subjectName string) (*unstructured.Unstructured, clierror.Error) {
	builder = builder.ForNamespace(cfg.namespace)

	if cfg.role != "" {
		builder = builder.
			ForRole(cfg.role).
			ForBindingName(fmt.Sprintf("%s-%s-binding", cfg.role, subjectName))
	} else {
		builder = builder.
			ForClusterRole(cfg.clusterrole).
			ForBindingName(fmt.Sprintf("%s-%s-binding", cfg.clusterrole, subjectName))
	}

	rbacResource, err := builder.BuildRoleBinding()
	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func applyRBAC(cfg *authorizeConfig, rbacResource *unstructured.Unstructured) clierror.Error {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	applier := authorization.NewResourceApplier(kubeClient)

	return applier.ApplyRBAC(cfg.Ctx, rbacResource)
}
