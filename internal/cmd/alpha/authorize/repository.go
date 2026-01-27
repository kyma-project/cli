package authorize

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type authorizeRepositoryConfig struct {
	*cmdcommon.KymaConfig
	repository    string
	clientId      string
	issuerURL     string
	prefix        string
	namespace     string
	clusterWide   bool
	role          string
	clusterrole   string
	name          string
	dryRun        bool
	requiredClaim map[string]string
	outputFormat  types.Format
}

func NewAuthorizeRepositoryCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := authorizeRepositoryConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "repository [flags]",
		Short: "Configures a trust between a Kyma cluster and a GitHub repository",
		Long:  "Configures a trust between a Kyma cluster and a GitHub repository by creating an OpenIDConnect resource and RoleBinding or ClusterRoleBinding",
		Example: `  # Authorize a repository with a namespaced Role (RoleBinding)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev

  # Authorize a repository cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole kyma-read-all --cluster-wide

  # Bind a repository to a ClusterRole within a single namespace (RoleBinding referencing ClusterRole)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole edit --namespace staging

  # Preview (dry-run) the YAML without applying
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role operator --namespace ops --dry-run -o yaml

  # Provide a custom OpenIDConnect resource name and username prefix
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole kyma-admin --cluster-wide --name custom-oidc --prefix gh-oidc:

  # Add additional required claims to the OIDC resource
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev --required-claim environment=dev --required-claim workflow=main

  # Use JSON output to inspect resources before apply
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev --dry-run -o json`,
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("repository"),
				flags.MarkRequired("client-id"),
				flags.MarkExactlyOneRequired("role", "clusterrole"),
				flags.MarkPrerequisites("role", "namespace"),
				flags.MarkPrerequisites("cluster-wide", "clusterrole"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorizeRepository(&cfg))
		},
	}

	// Required flags
	cmd.Flags().StringVar(&cfg.repository, "repository", "", "GitHub repo in owner/name format (e.g., kyma-project/cli) (required)")
	cmd.Flags().StringVar(&cfg.clientId, "client-id", "", "OIDC client ID (audience) expected in the token (required)")

	// Optional flags with defaults
	cmd.Flags().StringVar(&cfg.issuerURL, "issuer-url", "https://token.actions.githubusercontent.com", "OIDC issuer")
	cmd.Flags().StringVar(&cfg.prefix, "prefix", "", "Username prefix for the repository claim (e.g., gh-oidc:)")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace where the RoleBinding is created (required unless --cluster-wide is set)")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "If true, creates a ClusterRoleBinding; otherwise, a RoleBinding")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (namespaced)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (usable for RoleBinding and ClusterRoleBinding)")
	cmd.Flags().StringVar(&cfg.name, "name", "", "Name for the OpenIDConnect resource (optional; default derives from clientId)")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Prints resources without applying")
	cmd.Flags().StringToStringVar(&cfg.requiredClaim, "required-claim", nil, "Additional required claims (key=value) for the OpenIDConnect resource")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (yaml or json)")

	return cmd
}

func authorizeRepository(cfg *authorizeRepositoryConfig) clierror.Error {
	repositoryOIDCBuilder := authorization.NewOIDCBuilder(cfg.clientId, cfg.issuerURL).
		ForRepository(cfg.repository).
		ForRequiredClaims(cfg.requiredClaim).
		ForName(cfg.name).
		ForPrefix(cfg.prefix)
	oidcResource, err := repositoryOIDCBuilder.Build()

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build OIDC resource"))
	}

	rbacResource, rbacErr := buildRBACResourceForRepository(cfg, repositoryOIDCBuilder.GetUsernamePrefix())
	if rbacErr != nil {
		return rbacErr
	}

	msgSuffix := ""
	if cfg.dryRun {
		msgSuffix = " (dry run)"
	} else {
		if err := applyRepositoryAuthResources(cfg, oidcResource, rbacResource); err != nil {
			return err
		}
	}

	if cfg.outputFormat == "" {
		out.Msgfln("OpenIDConnect resource '%s' applied successfully.%s", oidcResource.GetName(), msgSuffix)
		out.Msgfln("%s '%s' applied successfully.%s", rbacResource.GetKind(), rbacResource.GetName(), msgSuffix)
	} else {
		return outputResources(cfg.outputFormat, []*unstructured.Unstructured{oidcResource, rbacResource})
	}

	return nil
}

func buildRBACResourceForRepository(cfg *authorizeRepositoryConfig, bindingPrefix string) (*unstructured.Unstructured, clierror.Error) {
	subjectKind, err := authorization.NewSubjectKindFrom("User")
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to generate binding"))
	}

	sanitizedRepoName := strings.ReplaceAll(cfg.repository, "/", "-")
	var bindingName string
	if cfg.role != "" {
		bindingName = fmt.Sprintf("%s-%s-binding", sanitizedRepoName, cfg.role)
	} else {
		bindingName = fmt.Sprintf("%s-%s-binding", sanitizedRepoName, cfg.clusterrole)
	}

	builder := authorization.NewRBACBuilder().
		ForPrefix(bindingPrefix).
		ForSubjectKind(subjectKind).
		ForSubjectName(cfg.repository).
		ForBindingName(bindingName)

	if cfg.clusterWide {
		return buildClusterRoleBindingForRepository(builder, cfg.clusterrole)
	}

	return buildRoleBindingForRepository(builder, cfg)
}

func buildClusterRoleBindingForRepository(builder *authorization.RBACBuilder, clusterrole string) (*unstructured.Unstructured, clierror.Error) {
	rbacResource, err := builder.ForClusterRole(clusterrole).BuildClusterRoleBinding()
	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func buildRoleBindingForRepository(builder *authorization.RBACBuilder, cfg *authorizeRepositoryConfig) (*unstructured.Unstructured, clierror.Error) {
	builder = builder.ForNamespace(cfg.namespace)

	if cfg.role != "" {
		builder = builder.ForRole(cfg.role)
	} else {
		builder = builder.ForClusterRole(cfg.clusterrole)
	}

	rbacResource, err := builder.BuildRoleBinding()
	if err != nil {
		return nil, err
	}
	return rbacResource, nil
}

func applyRepositoryAuthResources(cfg *authorizeRepositoryConfig, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	applier := authorization.NewResourceApplier(kubeClient)

	if err := applier.ApplyOIDC(cfg.Ctx, oidcResource); err != nil {
		return err
	}

	return applier.ApplyRBAC(cfg.Ctx, rbacResource)
}
