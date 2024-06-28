package oidc

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
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type oidcConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	cisCredentialsPath  string
	output              string
	audience            string
	token               string
	idTokenRequestURL   string
	idTokenRequestToken string
}

type TokenData struct {
	Count int    `json:"count"`
	Value string `json:"value"`
}

func NewOIDCCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := oidcConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "Create kubeconfig with an OIDC token",
		Long:  "Create kubeconfig with an OIDC token generated with a Github Actions token",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.complete())
			clierror.Check(cfg.validate())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runOIDC(&cfg))
		},
	}

	cfg.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&cfg.cisCredentialsPath, "credentials-path", "", "Path to the CIS credentials file.")
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the output kubeconfig file")

	cmd.Flags().StringVar(&cfg.token, "token", "", "Token used in the kubeconfig")
	cmd.Flags().StringVar(&cfg.audience, "audience", "", "Audience of the token")
	cmd.Flags().StringVar(&cfg.idTokenRequestURL, "id-token-request-url", "", "URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable")

	cmd.MarkFlagsOneRequired("kubeconfig", "credentials-path")
	cmd.MarkFlagsMutuallyExclusive("kubeconfig", "credentials-path")

	cmd.MarkFlagsMutuallyExclusive("token", "id-token-request-url")
	cmd.MarkFlagsMutuallyExclusive("token", "audience")

	return cmd
}

func (cfg *oidcConfig) complete() clierror.Error {
	if cfg.idTokenRequestURL == "" {
		cfg.idTokenRequestURL = os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	}
	cfg.idTokenRequestToken = os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")

	if cfg.cisCredentialsPath == "" {
		return cfg.KubeClientConfig.Complete()
	}
	return nil
}

func (cfg *oidcConfig) validate() clierror.Error {
	// if user explicitly provides token we don't need to run rest of the checks
	if cfg.token != "" {
		return nil
	}

	if cfg.idTokenRequestURL == "" {
		return clierror.New(
			"ID token request URL is required if --token is not provided",
			"make sure you're running the command in Github Actions environment",
			"provide id-token-request-url flag or ACTIONS_ID_TOKEN_REQUEST_URL env variable",
		)
	}

	if cfg.idTokenRequestToken == "" {
		return clierror.New(
			"ACTIONS_ID_TOKEN_REQUEST_TOKEN env variable is required if --token is not provided",
			"make sure you're running the command in Github Actions environment",
		)
	}
	return nil
}

func runOIDC(cfg *oidcConfig) clierror.Error {
	var clierr clierror.Error
	token := cfg.token
	if cfg.token == "" {
		// get Github token
		token, clierr = getGithubToken(cfg.idTokenRequestURL, cfg.idTokenRequestToken, cfg.audience)
		if clierr != nil {
			return clierror.WrapE(clierr, clierror.New("failed to get token"))
		}
	}

	var kubeconfig *api.Config

	if cfg.cisCredentialsPath != "" {
		kubeconfig, clierr = getKubeconfigFromCIS(cfg)
		if clierr != nil {
			return clierror.WrapE(clierr, clierror.New("failed to get kubeconfig from CIS"))
		}
	} else {
		kubeconfig = cfg.KubeClient.APIConfig()
	}

	enrichedKubeconfig := createKubeconfig(kubeconfig, token)

	err := kube.SaveConfig(enrichedKubeconfig, cfg.output)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to save kubeconfig"))
	}

	return nil
}

func getKubeconfigFromCIS(cfg *oidcConfig) (*api.Config, clierror.Error) {
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

	tokenData := TokenData{}
	err = json.NewDecoder(response.Body).Decode(&tokenData)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to decode token response"))
	}
	return tokenData.Value, nil
}

func createKubeconfig(kubeconfig *api.Config, token string) *api.Config {
	currentUser := kubeconfig.Contexts[kubeconfig.CurrentContext].AuthInfo
	config := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   kubeconfig.Clusters,
		AuthInfos: map[string]*api.AuthInfo{
			currentUser: {
				Token: token,
			},
		},
		Contexts:       kubeconfig.Contexts,
		CurrentContext: kubeconfig.CurrentContext,
		Extensions:     kubeconfig.Extensions,
		Preferences:    kubeconfig.Preferences,
	}

	return config
}
