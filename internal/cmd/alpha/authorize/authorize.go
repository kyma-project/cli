package authorize

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type authorizeConfig struct {
	*cmdcommon.KymaConfig

	authTarget              string // user, group, serviceaccount
	name                    []string
	namespace               string
	clusterWide             bool
	role                    string
	clusterrole             string
	serviceAccountNamespace string
	dryRun                  bool
	outputFormat            types.Format
}

func NewAuthorizeCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := authorizeConfig{
		KymaConfig: kymaConfig,
	}

	validAuthTargets := []string{"user", "group", "serviceaccount"}

	cmd := &cobra.Command{
		Use:   "authorize <authTarget> [flags]",
		Short: "Authorize a subject (user, group, or service account) with Kyma RBAC resources",
		Long:  "Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a user, group, or service account.",
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
				flags.MarkPrerequisites("cluster-wide", "clusterrole"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorize(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace for RoleBinding (required when binding a Role or binding a ClusterRole to a specific namespace)")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "Create a ClusterRoleBinding for cluster-wide access (requires --clusterrole)")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (creates RoleBinding in specified namespace)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (for ClusterRoleBinding with --cluster-wide, or RoleBinding in namespace)")
	cmd.Flags().StringSliceVar(&cfg.name, "name", []string{}, "Name(s) of the subject(s) to authorize (required)")
	cmd.Flags().StringVar(&cfg.serviceAccountNamespace, "sa-namespace", "", "Namespace for the service account subject. Defaults to the RoleBinding namespace when not specified.")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Preview the YAML/JSON output without applying resources to the cluster")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format for dry-run (yaml or json)")

	cmd.AddCommand(NewAuthorizeRepositoryCMD(kymaConfig))

	return cmd
}

func authorize(cfg *authorizeConfig) clierror.Error {
	rbacResources, err := buildRBACResources(cfg)
	if err != nil {
		return err
	}

	for _, res := range rbacResources {
		msgSuffix := ""
		if cfg.dryRun {
			msgSuffix = " (dry run)"
		} else {
			if applyErr := applyRBAC(cfg, res); applyErr != nil {
				return applyErr
			}
		}

		if cfg.outputFormat == "" {
			out.Msgfln("%s '%s' applied successfully.%s", res.GetKind(), res.GetName(), msgSuffix)
		}
	}

	if cfg.outputFormat != "" {
		clierr := outputResources(cfg.outputFormat, rbacResources)
		if clierr != nil {
			return clierr
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
		ForServiceAccountNamespace(cfg.serviceAccountNamespace).
		BuildClusterRoleBinding()

	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func buildRoleBinding(builder *authorization.RBACBuilder, cfg *authorizeConfig, subjectName string) (*unstructured.Unstructured, clierror.Error) {
	builder = builder.
		ForNamespace(cfg.namespace).
		ForServiceAccountNamespace(cfg.serviceAccountNamespace)

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
	clierr = applier.ApplyRBAC(cfg.Ctx, rbacResource)
	if clierr != nil {
		return clierr
	}

	return nil
}
