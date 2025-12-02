package kubeconfig

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/github"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type generateConfig struct {
	*cmdcommon.KymaConfig

	// common flags
	output string

	server string
	ca     string
	caData string

	// ServiceAccount-based flow flags
	serviceAccount string
	clusterRole    string
	role           string
	namespace      string
	time           string
	permanent      bool
	clusterWide    bool

	// OIDC flow options
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
		Example: `  # Generate a permanent access (kubeconfig) for a new or existing ServiceAccount 
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --permanent

  # Generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --permanent

  # Generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given Role
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --role <r_name> --permanent

  # Generate time-constrained access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a cluster-wide binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --cluster-wide --time 2h
  
  # Generate a kubeconfig with an OIDC token
  kyma alpha kubeconfig generate --token <token>

  # Generate a kubeconfig with an OIDC token based on a kubeconfig from the CIS
  kyma alpha kubeconfig generate --token <token> --credentials-path <cis_credentials>

  # Generate a kubeconfig with an requested OIDC token with audience option
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate --id-token-request-url <url> --audience <audience>

  # Generate a kubeconfig with an requested OIDC token with url from env
  export ACTIONS_ID_TOKEN_REQUEST_URL=<url>
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate`,
		Short:   "Generate kubeconfig with a Service Account-based or oidc tokens",
		Long:    "Use this command to generate kubeconfig file with a Service Account-based or oidc tokens",
		Aliases: []string{"gen"},
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfg.complete(cmd)
			allFlags := &pflag.FlagSet{}
			allFlags.AddFlagSet(cmd.Flags())
			allFlags.AddFlagSet(cmd.PersistentFlags())

			clierror.Check(flags.Validate(allFlags,
				flags.MarkOneRequired("serviceaccount", "token", "id-token-request-url", "oidc-name"),
				flags.MarkExclusive("token", "id-token-request-url", "audience"),
				flags.MarkExclusive("permanent", "time"),
				flags.MarkExclusive("cluster-wide", "role"),
				flags.MarkExclusive("kubeconfig", "server"),
				flags.MarkExclusive("kubeconfig", "certificate-authority"),
				// MarkRequired("server", ["certificate-authority", "certificate-authority-data"]) implemented in cfg.validate()
				flags.MarkExclusive("certificate-authority", "certificate-authority-data"),
				// providing server + ca without user data only makes sense for the token / id-token-request-url flow
				// other paths require access to cluster resources to create new kubeconfig
				flags.MarkExclusive("oidc-name", "server"),
				flags.MarkExclusive("serviceaccount", "server"),
			))
			clierror.Check(cfg.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runGenerate(cfg))
		},
	}

	// common
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the kubeconfig file output. If not provided, the kubeconfig is printed")

	// alternative to kubeconfig, used with token authorization
	cmd.Flags().StringVarP(&cfg.server, "server", "s", "", "The address and port of the Kubernetes API server")
	cmd.Flags().StringVar(&cfg.ca, "certificate-authority", "", "Path to a cert file for the certificate authority")
	cmd.Flags().StringVar(&cfg.caData, "certificate-authority-data", "", "Base64 encoded certificate authority data")

	// ServiceAccount-based flow
	cmd.Flags().StringVar(&cfg.serviceAccount, "serviceaccount", "", "Name of the Service Account (in the given namespace) to be used as a subject of the generated kubeconfig. If the Service Account does not exist, it is created")
	cmd.Flags().StringVar(&cfg.clusterRole, "clusterrole", "", "Name of the ClusterRole to bind the Service Account to (optional)")
	cmd.Flags().StringVar(&cfg.role, "role", "", "Name of the Role in the given namespace to bind the Service Account to (optional)")
	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "default", "Namespace in which the subject Service Account is to be found or is created")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "Determines how long the token is valid, by default 1h (use h for hours and d for days)")

	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Determines if the token is valid indefinitely")
	cmd.Flags().BoolVar(&cfg.clusterWide, "cluster-wide", false, "Determines if the binding to the ClusterRole is cluster-wide")

	// OIDC flow
	cmd.Flags().StringVar(&cfg.token, "token", "", "Token used in the kubeconfig")
	cmd.Flags().StringVar(&cfg.audience, "audience", "", "Audience of the token")
	cmd.Flags().StringVar(&cfg.idTokenRequestURL, "id-token-request-url", "", "URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable")

	// generate from OIDC custom resource
	cmd.Flags().StringVar(&cfg.oidcName, "oidc-name", "", "Name of the OIDC custom resource from which the kubeconfig is generated")
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
			"make sure you're running the command in the GitHub Actions environment",
		)
	}
	if cfg.server != "" && (cfg.ca == "" && cfg.caData == "") {
		return clierror.New(
			"--certificate-authority or --certificate-authority-data flag is required when --server flag is provided",
		)
	}
	return nil
}

func runGenerate(cfg *generateConfig) clierror.Error {
	var generateFunc func(*generateConfig, kube.Client) (*api.Config, clierror.Error)

	if cfg.serviceAccount != "" {
		generateFunc = generateWithServiceAccount
	} else if cfg.oidcName != "" {
		generateFunc = generateWithOpenIDConnectorCustomResource
	} else {
		generateFunc = generateWithToken
	}

	client, clierr := getKubeClient(cfg)
	if clierr != nil {
		return clierr
	}

	kubeconfig, clierr := generateFunc(cfg, client)
	if clierr != nil {
		return clierr
	}

	// Print or write to file
	return returnKubeconfig(cfg, kubeconfig)
}

func getKubeClient(cfg *generateConfig) (kube.Client, clierror.Error) {

	if cfg.server != "" && (cfg.ca != "" || cfg.caData != "") {
		ca, err := getCAData(cfg)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("failed to get certificate authority data"))
		}

		tmpKubeconfig := api.Config{
			APIVersion:     "v1",
			Kind:           "Config",
			CurrentContext: "default",
			Clusters: map[string]*api.Cluster{
				"default": {
					Server: cfg.server,
					// TODO: use CertificateAuthority for cfg.ca file?
					CertificateAuthorityData: ca,
				},
			},
			Contexts: map[string]*api.Context{
				"default": {
					Cluster:  "default",
					AuthInfo: "default",
				},
			},
			AuthInfos: map[string]*api.AuthInfo{
				"default": {},
			},
		}

		client, err := kube.NewClientForConfig(&tmpKubeconfig)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("failed to create kube client from server and ca"))
		}
		return client, nil
	}
	return cfg.GetKubeClientWithClierr()
}

func getCAData(cfg *generateConfig) ([]byte, error) {
	if cfg.ca != "" {
		return os.ReadFile(cfg.ca)
	}
	if cfg.caData != "" {
		return base64.StdEncoding.DecodeString(cfg.caData)
	}
	return nil, errors.New("no certificate authority data provided")
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
		out.Msgln(string(message))
	}

	return nil
}

func generateWithToken(cfg *generateConfig, client kube.Client) (*api.Config, clierror.Error) {
	var clierr clierror.Error
	token := cfg.token
	if cfg.token == "" {
		// Get Github token if not provided
		token, clierr = github.GetToken(cfg.idTokenRequestURL, cfg.idTokenRequestToken, cfg.audience)
		if clierr != nil {
			return nil, clierror.WrapE(clierr, clierror.New("failed to get token"))
		}
	}

	kubeconfigTemplate := client.APIConfig()

	// Prepare kubeconfig based on template and token
	return kubeconfig.PrepareWithToken(kubeconfigTemplate, token), nil
}

func generateWithServiceAccount(cfg *generateConfig, kubeClient kube.Client) (*api.Config, clierror.Error) {

	// Ensure ServiceAccount, Bindings and secret with token
	clierr := setupServiceAccountWithBindings(cfg, kubeClient)
	if clierr != nil {
		return nil, clierror.WrapE(clierr, clierror.New("failed to create k8s resources"))
	}

	// Fill kubeconfig
	return kubeconfig.Prepare(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.time, cfg.output, cfg.permanent)
}

func generateWithOpenIDConnectorCustomResource(cfg *generateConfig, client kube.Client) (*api.Config, clierror.Error) {
	return kubeconfig.PrepareFromOpenIDConnectorResource(cfg.Ctx, client, cfg.oidcName)
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
			return clierror.Wrap(err, clierror.New("failed to create Secret"))
		}
	}
	return nil
}
