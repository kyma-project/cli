package kubeconfig

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/github"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type generateConfig struct {
	*cmdcommon.KymaConfig

	// common flags
	output string

	// ServiceAccount-based flow flags
	serviceAccount string
	clusterRole    string
	role           string
	namespace      string
	time           string
	permanent      bool
	clusterWide    bool

	// OIDC flow options
	cisCredentialsPath  string
	token               string
	audience            string
	idTokenRequestURL   string
	idTokenRequestToken string

	// OIDC CR
	oidcName string
}

func newGenerateCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := &generateConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use: "generate",
		Example: `# generate a permanent access (kubeconfig) for a new or existing ServiceAccount 
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --permanent

# generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --permanent

# generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given Role
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --role <r_name> --permanent

# generate time-constrained access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a cluster-wide binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --cluster-wide --time 2h
  
# generate a kubeconfig with an OIDC token
  kyma alpha kubeconfig generate --token <token>

# generate a kubeconfig with an OIDC token based on a kubeconfig from the CIS
  kyma alpha kubeconfig generate --token <token> --credentials-path <cis_credentials>

# generate a kubeconfig with an requested OIDC token with audience option
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate --id-token-request-url <url> --audience <audience>

# generate a kubeconfig with an requested OIDC token with url from env
  export ACTIONS_ID_TOKEN_REQUEST_URL=<url>
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate`,
		Short:   "Generate kubeconfig with a Service Account-based or oidc tokens",
		Long:    "Use this command to generate kubeconfig file with a Service Account-based or oidc tokens",
		Aliases: []string{"gen"},
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfg.complete(cmd)
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkOneRequired("serviceaccount", "token", "id-token-request-url", "oidc-name"),
				flags.MarkExclusive("token", "id-token-request-url", "audience"),
				flags.MarkExclusive("permanent", "time"),
				flags.MarkExclusive("cluster-wide", "role"),
			))
			clierror.Check(cfg.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.validate())
			clierror.Check(runGenerate(cfg))
		},
	}

	// common
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the kubeconfig file output. If not provided, the kubeconfig will be printed")

	// ServiceAccount-based flow
	cmd.Flags().StringVar(&cfg.serviceAccount, "serviceaccount", "", "Name of the Service Account (in the given Namespace) to be used as a subject of the generated kubeconfig. If the Service Account does not exist, it will be created")
	cmd.Flags().StringVar(&cfg.clusterRole, "clusterrole", "", "Name of the Cluster Role to bind the Service Account to (optional)")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Name of the Role in the given Namespace to bind the Service Account to (optional)")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace in which the subject Service Account is to be found or will be created")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "Determines how long the token should be valid, by default 1h (use h for hours and d for days)")

	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Determines if the token is valid indefinitely")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "Determines if the binding to the ClusterRole is cluster-wide")

	// OIDC flow
	cmd.Flags().StringVar(&cfg.cisCredentialsPath, "credentials-path", "", "Path to the CIS credentials file")
	cmd.Flags().StringVar(&cfg.token, "token", "", "Token used in the kubeconfig")
	cmd.Flags().StringVar(&cfg.audience, "audience", "", "Audience of the token")
	cmd.Flags().StringVar(&cfg.idTokenRequestURL, "id-token-request-url", "", "URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable")

	// generate from OIDC custom resource
	cmd.Flags().StringVar(&cfg.oidcName, "oidc-name", "", "Name of the OIDC Custom Resource from which the kubeconfig will be generated")
	return cmd
}

func (cfg *generateConfig) complete(cmd *cobra.Command) {
	// complete for OIDC flow
	requestUrlEnv := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if cfg.idTokenRequestURL == "" && requestUrlEnv != "" && cfg.token == "" {
		_ = cmd.Flags().Set("id-token-request-url", requestUrlEnv)
	}
	cfg.idTokenRequestToken = os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
}

func (cfg *generateConfig) validate() clierror.Error {
	if cfg.idTokenRequestURL != "" && cfg.idTokenRequestToken == "" {
		// check if request token is provided
		return clierror.New(
			"ACTIONS_ID_TOKEN_REQUEST_TOKEN env variable is required if --id-token-request-url flag or ACTIONS_ID_TOKEN_REQUEST_URL env were provided",
			"make sure you're running the command in Github Actions environment",
		)
	}
	return nil
}

func runGenerate(cfg *generateConfig) clierror.Error {
	var generateFunc func(*generateConfig) (*api.Config, clierror.Error)

	if cfg.serviceAccount != "" {
		generateFunc = generateWithServiceAccount
	} else if cfg.oidcName != "" {
		generateFunc = generateWithOpenIDConnectorCustomResource
	} else {
		generateFunc = generateWithToken
	}

	kubeconfig, clierr := generateFunc(cfg)
	if clierr != nil {
		return clierr
	}

	// Print or write to file
	return returnKubeconfig(cfg, kubeconfig)
}

func returnKubeconfig(cfg *generateConfig, kubeconfig *api.Config) clierror.Error {
	if cfg.output != "" {
		err := kube.SaveConfig(kubeconfig, cfg.output)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to save kubeconfig"))
		}
	} else {
		message, err := clientcmd.Write(*kubeconfig)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to print kubeconfig"))
		}
		fmt.Println(string(message))
	}

	return nil
}

func generateWithToken(cfg *generateConfig) (*api.Config, clierror.Error) {
	var clierr clierror.Error
	token := cfg.token
	if cfg.token == "" {
		// Get Github token if not provided
		token, clierr = github.GetToken(cfg.idTokenRequestURL, cfg.idTokenRequestToken, cfg.audience)
		if clierr != nil {
			return nil, clierror.WrapE(clierr, clierror.New("failed to get token"))
		}
	}

	var kubeconfigTemplate *api.Config
	if cfg.cisCredentialsPath != "" {
		// Get cluster kubeconfig from CIS
		kubeconfigTemplate, clierr = kubeconfig.GetFromCIS(cfg.cisCredentialsPath)
		if clierr != nil {
			return nil, clierror.WrapE(clierr, clierror.New("failed to get kubeconfig template for Kyma environment"))
		}
	} else {
		// Get cluster kubeconfig from cluster
		client, err := cfg.GetKubeClientWithClierr()
		if err != nil {
			return nil, err
		}

		kubeconfigTemplate = client.APIConfig()
	}

	// Prepare kubeconfig based on template and token
	return kubeconfig.PrepareWithToken(kubeconfigTemplate, token), nil
}

func generateWithServiceAccount(cfg *generateConfig) (*api.Config, clierror.Error) {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return nil, clierr
	}

	// Ensure ServiceAccount, Bindings and secret with token
	clierr = setupServiceAccountWithBindings(cfg, kubeClient)
	if clierr != nil {
		return nil, clierror.WrapE(clierr, clierror.New("failed to create k8s resources"))
	}

	// Fill kubeconfig
	return kubeconfig.Prepare(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.time, cfg.output, cfg.permanent)
}

func generateWithOpenIDConnectorCustomResource(cfg *generateConfig) (*api.Config, clierror.Error) {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return nil, clierr
	}

	return kubeconfig.PrepareFromOpenIDConnectorResource(cfg.Ctx, kubeClient, cfg.oidcName)
}

func setupServiceAccountWithBindings(cfg *generateConfig, kubeClient kube.Client) clierror.Error {
	// Get or Create Service Account
	err := resources.EnsureServiceAccount(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Service Account"))
	}

	// Create Role or ClusterRole Binding for the Service Account
	if cfg.clusterWide && cfg.clusterRole != "" {
		// Create ClusterRoleBinding for the Service Account
		err = resources.CreateClusterRoleBinding(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.clusterRole)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create Cluster Role Binding"))
		}
	}

	if !cfg.clusterWide && cfg.role != "" {
		// Create Role Binding to Role for the Service Account
		err = resources.CreateRoleBindingToRole(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.role)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create Role Binding"))
		}
	}

	if !cfg.clusterWide && cfg.clusterRole != "" {
		// Create Role Binding to Cluster Role for the Service Account
		err = resources.CreateRoleBindingToClusterRole(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.clusterRole)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create Role Binding"))
		}
	}

	// Create a service-account-token type secret
	if cfg.permanent {
		err = resources.CreateServiceAccountToken(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create secret"))
		}
	}
	return nil
}
