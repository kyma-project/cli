package authorize

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type authorizeConfig struct {
	*cmdcommon.KymaConfig
	repository   string
	clientId     string
	issuerURL    string
	prefix       string
	namespace    string
	clusterWide  bool
	role         string
	clusterrole  string
	name         string
	dryRun       bool
	outputFormat types.Format
}

func NewAuthorizeCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := authorizeConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "authorize repository [flags]",
		Short: "Configure trust between a Kyma cluster and a GitHub repository",
		Long:  "Configure trust between a Kyma cluster and a GitHub repository by creating an OpenIDConnect resource and RoleBinding or ClusterRoleBinding",
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("repository"),
				flags.MarkRequired("clientId"),
				flags.MarkExactlyOneRequired("role", "clusterrole"),
				flags.MarkPrerequisites("role", "namespace"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(authorize(&cfg))
		},
	}

	// Required flags
	cmd.Flags().StringVar(&cfg.repository, "repository", "", "GitHub repo in owner/name format (e.g., kyma-project/cli) (required)")
	cmd.Flags().StringVar(&cfg.clientId, "clientId", "", "OIDC client ID (audience) expected in the token (required)")

	// Optional flags with defaults
	cmd.Flags().StringVar(&cfg.issuerURL, "issuerURL", "https://token.actions.githubusercontent.com", "OIDC issuer")
	cmd.Flags().StringVar(&cfg.prefix, "prefix", "", "Username prefix for the repository claim (e.g., gh-oidc:)")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace for RoleBinding (required if not cluster-wide and binding a Role or namespaced ClusterRole)")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "If true, create a ClusterRoleBinding; otherwise, a RoleBinding")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Role name to bind (namespaced)")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "ClusterRole name to bind (usable for RoleBinding or ClusterRoleBinding)")
	cmd.Flags().StringVar(&cfg.name, "name", "", "Name for the OpenIDConnect resource (optional; default derives from clientId)")
	cmd.Flags().BoolVar(&cfg.dryRun, "dry-run", false, "Print resources without applying")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (yaml or json)")

	return cmd
}

func authorize(cfg *authorizeConfig) clierror.Error {
	oidcResource, err := buildOIDCResource(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build OIDC resource"))
	}

	rbacResource, err := buildRBACResource(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to build RBAC resource"))
	}

	// Handle dry-run output
	if cfg.dryRun {
		return outputResources(cfg, oidcResource, rbacResource)
	}

	// Apply resources to cluster
	return applyResources(cfg, oidcResource, rbacResource)
}

func buildOIDCResource(cfg *authorizeConfig) (*unstructured.Unstructured, error) {
	repositoryOIDCBuilder := authorization.NewOIDCBuilder(cfg.clientId, cfg.issuerURL)
	oidcResource, err := repositoryOIDCBuilder.
		ForRepository(cfg.repository).
		ForName(cfg.name).
		Build()

	if err != nil {
		return nil, err
	}

	return oidcResource, nil
}

func buildRBACResource(cfg *authorizeConfig) (*unstructured.Unstructured, error) {
	builder := authorization.NewRBACBuilder().
		ForRepository(cfg.repository).
		ForPrefix(cfg.prefix)

	if cfg.clusterWide {
		return buildClusterRoleBinding(builder, cfg.clusterrole)
	}

	return buildRoleBinding(builder, cfg)
}

func buildClusterRoleBinding(builder *authorization.RBACBuilder, clusterrole string) (*unstructured.Unstructured, error) {
	rbacResource, err := builder.ForClusterRole(clusterrole).BuildClusterRoleBinding()
	if err != nil {
		return nil, fmt.Errorf("failed to build ClusterRoleBinding: %w", err)
	}
	return rbacResource, nil
}

func buildRoleBinding(builder *authorization.RBACBuilder, cfg *authorizeConfig) (*unstructured.Unstructured, error) {
	builder = builder.ForNamespace(cfg.namespace)

	if cfg.role != "" {
		builder = builder.ForRole(cfg.role)
	} else {
		builder = builder.ForClusterRole(cfg.clusterrole)
	}

	rbacResource, err := builder.BuildRoleBinding()
	if err != nil {
		return nil, fmt.Errorf("failed to build RoleBinding: %w", err)
	}
	return rbacResource, nil
}

func outputResources(cfg *authorizeConfig, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	printer := out.Default

	switch cfg.outputFormat {
	case types.JSONFormat:
		return outputJSON(printer, oidcResource, rbacResource)
	case types.YAMLFormat, types.DefaultFormat:
		return outputYAML(printer, oidcResource, rbacResource)
	default:
		return outputYAML(printer, oidcResource, rbacResource)
	}
}

func outputJSON(printer *out.Printer, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	resources := []any{oidcResource.Object, rbacResource.Object}

	obj, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal resources to JSON"))
	}

	printer.Msgln(string(obj))
	return nil
}

func outputYAML(printer *out.Printer, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	// Output OpenIDConnect resource
	oidcBytes, err := yaml.Marshal(oidcResource.Object)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal OpenIDConnect resource to YAML"))
	}

	printer.Msgln(string(oidcBytes))

	// Add separator
	printer.Msgln("---")

	// Output RBAC resource directly (already clean unstructured format)
	rbacBytes, err := yaml.Marshal(rbacResource.Object)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to marshal RBAC resource to YAML"))
	}

	printer.Msgln(string(rbacBytes))

	return nil
}

func applyResources(cfg *authorizeConfig, oidcResource *unstructured.Unstructured, rbacResource *unstructured.Unstructured) clierror.Error {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	applier := authorization.NewResourceApplier(kubeClient)
	return applier.ApplyResources(cfg.Ctx, oidcResource, rbacResource)
}
