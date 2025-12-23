package authorize

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type authorizeConfig struct {
	*cmdcommon.KymaConfig

	authTarget              string // user, group, serviceaccount
	name                    []string
	bindingName             string
	namespace               string
	clusterWide             bool
	role                    string
	clusterrole             string
	serviceAccountNamespace string
	dryRun                  bool
	force                   bool
	outputFormat            types.Format
}

func NewAuthorizeCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorize",
		Short: "Authorizes a subject (user, group, or service account) with Kyma RBAC resources",
		Long:  "Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a user, group, or service account.",
	}

	cmd.AddCommand(NewAuthorizeUserCMD(kymaConfig))
	cmd.AddCommand(NewAuthorizeGroupCMD(kymaConfig))
	cmd.AddCommand(NewAuthorizeServiceAccountCMD(kymaConfig))
	cmd.AddCommand(NewAuthorizeRepositoryCMD(kymaConfig))

	return cmd
}

func NewAuthorizeUserCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	return newAuthorizeSubjectCMD(kymaConfig, "user", "User", `  # Bind a user to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice --role view --namespace dev

  # Bind multiple users to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice,bob,james --role view --namespace dev

  # Bind a user cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize user --name ci-bot --clusterrole kyma-admin --cluster-wide

  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize user --name bob --role operator --namespace ops --dry-run -o yaml`)
}

func NewAuthorizeGroupCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	return newAuthorizeSubjectCMD(kymaConfig, "group", "Group", `  # Bind a group cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize group --name team-observability --clusterrole kyma-read-all --cluster-wide

  # Bind a group to a namespaced Role (RoleBinding)
  kyma alpha authorize group --name developers --role edit --namespace dev

  # Generate JSON for a cluster-wide binding
  kyma alpha authorize group --name ops-team --clusterrole cluster-admin --cluster-wide -o json
  
  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize group --name ops-team --role edit --namespace dev --dry-run -o yaml`)
}

func NewAuthorizeServiceAccountCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := authorizeConfig{
		KymaConfig: kymaConfig,
		authTarget: "serviceaccount",
	}

	cmd := &cobra.Command{
		Use:   "serviceaccount",
		Short: "Authorizes a service account with Kyma RBAC resources",
		Long:  "Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a ServiceAccount.",
		Example: `  # Bind a service account to a ClusterRole within a namespace (RoleBinding referencing a ClusterRole)
  kyma alpha authorize serviceaccount --name deployer-sa --clusterrole edit --namespace staging

  # Bind a service account cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize serviceaccount --name system-bot --clusterrole cluster-admin --cluster-wide

  # Specify a different namespace for the service account subject
  kyma alpha authorize serviceaccount --name remote-sa --sa-namespace tools --clusterrole view --namespace dev
   
  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize serviceaccount --name remote-sa --role edit --namespace dev --dry-run -o yaml 
  `,
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("name"),
				flags.MarkExactlyOneRequired("role", "clusterrole"),
				flags.MarkPrerequisites("role", "namespace"),
				flags.MarkPrerequisites("cluster-wide", "clusterrole"),
				flags.MarkExclusive("namespace", "cluster-wide"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorize(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace for RoleBinding (required when binding a Role or binding a ClusterRole to a specific namespace)")

	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "Creates a ClusterRoleBinding for cluster-wide access (requires --clusterrole)")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (creates RoleBinding in specified namespace)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (for ClusterRoleBinding with --cluster-wide, or RoleBinding in namespace)")
	cmd.Flags().StringSliceVar(&cfg.name, "name", []string{}, "Name(s) of the subject(s) to authorize (required)")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Previews the YAML/JSON output without applying resources to the cluster")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Forces application of the binding, overwriting if it already exists")
	cmd.Flags().StringVar(&cfg.bindingName, "binding-name", "", "Custom name for the RoleBinding or ClusterRoleBinding. If not specified, a name is auto-generated based on the role and subject")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format for dry-run (yaml or json)")
	cmd.Flags().StringVar(&cfg.serviceAccountNamespace, "sa-namespace", "", "Namespace for the service account subject. Defaults to the RoleBinding namespace when not specified.")

	return cmd
}

func newAuthorizeSubjectCMD(kymaConfig *cmdcommon.KymaConfig, authTarget, subjectType, examples string) *cobra.Command {
	cfg := authorizeConfig{
		KymaConfig: kymaConfig,
		authTarget: authTarget,
	}

	cmd := &cobra.Command{
		Use:     authTarget,
		Short:   fmt.Sprintf("Authorizes a %s with Kyma RBAC resources", strings.ToLower(subjectType)),
		Long:    fmt.Sprintf("Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a %s.", strings.ToLower(subjectType)),
		Example: examples,
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("name"),
				flags.MarkExactlyOneRequired("role", "clusterrole"),
				flags.MarkPrerequisites("role", "namespace"),
				flags.MarkPrerequisites("cluster-wide", "clusterrole"),
				flags.MarkExclusive("namespace", "cluster-wide"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorize(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace for RoleBinding (required when binding a Role or binding a ClusterRole to a specific namespace)")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "Creates a ClusterRoleBinding for cluster-wide access (requires --clusterrole)")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (creates RoleBinding in specified namespace)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (for ClusterRoleBinding with --cluster-wide, or RoleBinding in namespace)")
	cmd.Flags().StringSliceVar(&cfg.name, "name", []string{}, "Name(s) of the subject(s) to authorize (required)")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Previews the YAML/JSON output without applying resources to the cluster")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Forces application of the binding, overwriting if it already exists")
	cmd.Flags().StringVar(&cfg.bindingName, "binding-name", "", "Custom name for the RoleBinding or ClusterRoleBinding. If not specified, a name is auto-generated based on the role and subject")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format for dry-run (yaml or json)")

	return cmd
}

func authorize(cfg *authorizeConfig) clierror.Error {
	rbacResources, err := buildRBACResources(cfg)
	if err != nil {
		return err
	}

	err = validateResourcesNamespace(cfg)
	if err != nil {
		return err
	}

	for _, res := range rbacResources {
		if cfg.dryRun {
			printSuccessMessage(res, " (dry run)", cfg.outputFormat)
			continue
		}

		rbacApplied, applyErr := applyRBAC(cfg, res)
		if applyErr != nil {
			return applyErr
		}

		if rbacApplied {
			printSuccessMessage(res, "", cfg.outputFormat)
		}
	}

	if cfg.outputFormat != "" {
		return outputResources(cfg.outputFormat, rbacResources)
	}

	return nil
}

func validateResourcesNamespace(cfg *authorizeConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := resources.NamespaceExists(cfg.Ctx, client, cfg.namespace)
	if err != nil {
		return clierror.Wrap(err, clierror.New("namespace validation failed"))
	}

	return nil
}

func printSuccessMessage(res *unstructured.Unstructured, suffix string, outputFormat types.Format) {
	if outputFormat == "" {
		out.Msgfln("%s '%s' applied successfully.%s", res.GetKind(), res.GetName(), suffix)
	}
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
		ForBindingName(getBindingName(cfg, subjectName)).
		ForServiceAccountNamespace(cfg.serviceAccountNamespace).
		BuildClusterRoleBinding()

	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func getBindingName(cfg *authorizeConfig, subjectName string) string {
	var role string

	if cfg.clusterrole != "" {
		role = cfg.clusterrole
	} else {
		role = cfg.role
	}

	var bindingName string

	if cfg.bindingName == "" {
		bindingName = fmt.Sprintf("%s-%s-%s", role, subjectName, cfg.authTarget)
	} else {
		bindingName = cfg.bindingName
	}

	return bindingName
}

func buildRoleBinding(builder *authorization.RBACBuilder, cfg *authorizeConfig, subjectName string) (*unstructured.Unstructured, clierror.Error) {
	builder = builder.
		ForNamespace(cfg.namespace).
		ForServiceAccountNamespace(cfg.serviceAccountNamespace)

	if cfg.role != "" {
		builder = builder.
			ForRole(cfg.role).
			ForBindingName(getBindingName(cfg, subjectName))
	} else {
		builder = builder.
			ForClusterRole(cfg.clusterrole).
			ForBindingName(getBindingName(cfg, subjectName))
	}

	rbacResource, err := builder.BuildRoleBinding()
	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func applyRBAC(cfg *authorizeConfig, rbacResource *unstructured.Unstructured) (bool, clierror.Error) {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return false, clierr
	}

	if resources.BindingExists(cfg.Ctx, kubeClient, rbacResource) && !cfg.force {
		promptMsg := fmt.Sprintf("Binding with name %s already exists. Do you want to override it?", rbacResource.GetName())
		overridePrompt := prompt.NewBool(promptMsg, false)
		promptResult, err := overridePrompt.Prompt()
		if err != nil {
			return false, clierror.Wrap(err, clierror.New("failed to prompt"))
		}

		if !promptResult {
			return false, nil
		}
	}

	applier := authorization.NewResourceApplier(kubeClient)
	clierr = applier.ApplyRBAC(cfg.Ctx, rbacResource)
	if clierr != nil {
		return false, clierr
	}

	return true, nil
}
