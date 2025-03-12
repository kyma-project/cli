package kubeconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/cis"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/flags"
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
	namespace      string
	time           string
	permanent      bool

	// OIDC flow options
	cisCredentialsPath  string
	token               string
	audience            string
	idTokenRequestURL   string
	idTokenRequestToken string
}

func newGenerateCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := &generateConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use: "generate",
		Example: `# generate a kubeconfig with a ServiceAccount-based token and certificate
  kyma@v3 alpha kubeconfig generate --serviceaccount <sa_name> --clusterrole <cr_name> --namespace <ns_name> --permanent

# generate a kubeconfig with an OIDC token
  kyma@v3 alpha kubeconfig generate --token <token>

# generate a kubeconfig with an OIDC token based on a kubeconfig from the CIS
  kyma@v3 alpha kubeconfig generate --token <token> --credentials-path <cis_credentials>

# generate a kubeconfig with an requested OIDC token with audience option
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma@v3 alpha kubeconfig generate --id-token-request-url <url> --audience <audience>

# generate a kubeconfig with an requested OIDC token with url from env
  export ACTIONS_ID_TOKEN_REQUEST_URL=<url>
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma@v3 alpha kubeconfig generate`,
		Short:   "Generate kubeconfig with a Service Account-based or oidc tokens",
		Long:    "Use this command to generate kubeconfig file with a Service Account-based or oidc tokens",
		Aliases: []string{"gen"},
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfg.complete(cmd)
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkOneRequired("serviceaccount", "token", "id-token-request-url"),
				flags.MarkRequiredTogether("serviceaccount", "clusterrole"),
				flags.MarkExclusive("token", "id-token-request-url", "audience"),
			))
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.validate())
			clierror.Check(runGenerate(cfg))
		},
	}

	// common
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the kubeconfig file output. If not provided, the kubeconfig will be printed")

	// ServiceAccount-based flow
	cmd.Flags().StringVar(&cfg.serviceAccount, "serviceaccount", "", "Name of the Service Account to be created")
	cmd.Flags().StringVar(&cfg.clusterRole, "clusterrole", "", "Name of the Cluster Role to bind the Service Account to")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace in which the resource is created")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "Determines how long the token should be valid, by default 1h (use h for hours and d for days)")
	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Determines if the token is valid indefinitely")

	// OIDC flow
	cmd.Flags().StringVar(&cfg.cisCredentialsPath, "credentials-path", "", "Path to the CIS credentials file")
	cmd.Flags().StringVar(&cfg.token, "token", "", "Token used in the kubeconfig")
	cmd.Flags().StringVar(&cfg.audience, "audience", "", "Audience of the token")
	cmd.Flags().StringVar(&cfg.idTokenRequestURL, "id-token-request-url", "", "URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable")

	return cmd
}

func (cfg *generateConfig) complete(cmd *cobra.Command) {
	// complete for OIDC flow
	requestUrlEnv := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if cfg.idTokenRequestURL == "" && requestUrlEnv != "" {
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
	generateFunc := generateWithServiceAccount
	if cfg.serviceAccount == "" {
		// change to OIDC flow if serviceaccount flag is not set
		generateFunc = generateWithToken
	}

	kubeconfig, clierr := generateFunc(cfg)
	if clierr != nil {
		return clierr
	}

	return returnKubeconfig(cfg, kubeconfig)
}

func returnKubeconfig(cfg *generateConfig, kubeconfig *api.Config) clierror.Error {
	if cfg.output != "" {
		// Print or write to file
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
		// get Github token
		token, clierr = getGithubToken(cfg.idTokenRequestURL, cfg.idTokenRequestToken, cfg.audience)
		if clierr != nil {
			return nil, clierror.WrapE(clierr, clierror.New("failed to get token"))
		}
	}

	var kubeconfig *api.Config

	if cfg.cisCredentialsPath != "" {
		kubeconfig, clierr = getKubeconfigFromCIS(cfg)
		if clierr != nil {
			return nil, clierror.WrapE(clierr, clierror.New("failed to get kubeconfig from CIS"))
		}
	} else {
		client, err := cfg.GetKubeClientWithClierr()
		if err != nil {
			return nil, err
		}

		kubeconfig = client.APIConfig()
	}

	return createKubeconfigWithToken(kubeconfig, token), nil
}

func getKubeconfigFromCIS(cfg *generateConfig) (*api.Config, clierror.Error) {
	// TODO: maybe refactor with provision command to not duplicate localCISClient provisioning
	credentials, err := auth.LoadCISCredentials(cfg.cisCredentialsPath)
	if err != nil {
		return nil, err
	}
	token, err := auth.GetOAuthToken(
		credentials.GrantType,
		credentials.UAA.URL,
		credentials.UAA.ClientID,
		credentials.UAA.ClientSecret,
	)
	if err != nil {
		var hints []string
		if strings.Contains(err.String(), "Internal Server Error") {
			hints = append(hints, "check if CIS grant type is set to client credentials")
		}

		return nil, clierror.WrapE(err, clierror.New("failed to get access token", hints...))
	}

	localCISClient := cis.NewLocalClient(credentials, token)
	kubeconfigString, err := localCISClient.GetKymaKubeconfig()
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("failed to get kubeconfig"))
	}

	kubeconfig, err := parseKubeconfig(kubeconfigString)
	if err != nil {
		return nil, clierror.WrapE(err, clierror.New("failed to parse kubeconfig"))
	}
	return kubeconfig, nil
}

func parseKubeconfig(kubeconfigString string) (*api.Config, clierror.Error) {
	kubeconfig, err := clientcmd.Load([]byte(kubeconfigString))
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to parse kubeconfig string"))
	}
	return kubeconfig, nil
}

func getGithubToken(url, requestToken, audience string) (string, clierror.Error) {
	// create http client

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to create request"))
	}
	if audience != "" {
		q := request.URL.Query()
		q.Add("audience", audience)
		request.URL.RawQuery = q.Encode()
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", requestToken))
	request.Header.Add("Accept", "application/json; api-version=2.0")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to get token from Github"))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", clierror.New(fmt.Sprintf("Invalid server response: %d", response.StatusCode))
	}

	tokenData := struct {
		Count int    `json:"count"`
		Value string `json:"value"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&tokenData)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to decode token response"))
	}
	return tokenData.Value, nil
}

func createKubeconfigWithToken(kubeconfigBase *api.Config, token string) *api.Config {
	currentUser := kubeconfigBase.Contexts[kubeconfigBase.CurrentContext].AuthInfo
	config := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   kubeconfigBase.Clusters,
		AuthInfos: map[string]*api.AuthInfo{
			currentUser: {
				Token: token,
			},
		},
		Contexts:       kubeconfigBase.Contexts,
		CurrentContext: kubeconfigBase.CurrentContext,
		Extensions:     kubeconfigBase.Extensions,
		Preferences:    kubeconfigBase.Preferences,
	}

	return config
}

func generateWithServiceAccount(cfg *generateConfig) (*api.Config, clierror.Error) {
	kubeClient, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return nil, clierr
	}

	// Create ServiceAccount, ClusterRoleBinding and secret with token
	clierr = registerServiceAccount(cfg, kubeClient)
	if clierr != nil {
		return nil, clierror.WrapE(clierr, clierror.New("failed to create objects"))
	}

	// Fill kubeconfig
	return kubeconfig.Prepare(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.time, cfg.output, cfg.permanent)
}

func registerServiceAccount(cfg *generateConfig, kubeClient kube.Client) clierror.Error {
	// Create Service Account
	err := resources.CreateServiceAccount(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Service Account"))
	}
	// Create Role Binding for the Service Account
	err = resources.CreateClusterRoleBinding(cfg.Ctx, kubeClient, cfg.serviceAccount, cfg.namespace, cfg.clusterRole)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Cluster Role Binding"))
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
