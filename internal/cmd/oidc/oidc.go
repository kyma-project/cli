package oidc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type oidcConfig struct {
	*cmdcommon.KymaConfig

	output              string
	caCertificate       string
	clusterServer       string
	audience            string
	IDTokenRequestURL   string
	IDTokenRequestToken string
}

type TokenData struct {
	Count int    `json:"count"`
	Value string `json:"value"`
}

func NewOIDCCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := oidcConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "Create kubeconfig with an OIDC token",
		Long:  "Create kubeconfig with an OIDC token generated with a Github Actions token",
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.complete())
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(runOIDC(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the output kubeconfig file")
	cmd.Flags().StringVar(&cfg.caCertificate, "ca-certificate", "", "Path to the CA certificate file")
	cmd.Flags().StringVar(&cfg.clusterServer, "cluster-server", "", "URL of the cluster server")
	cmd.Flags().StringVar(&cfg.audience, "audience", "", "Audience of the token")
	cmd.Flags().StringVar(&cfg.IDTokenRequestURL, "id-token-request-url", "", "URL to request the ID token")

	_ = cmd.MarkFlagRequired("ca-certificate")
	_ = cmd.MarkFlagRequired("cluster-server")

	return cmd
}

func (cfg *oidcConfig) complete() clierror.Error {
	if cfg.IDTokenRequestURL == "" {
		cfg.IDTokenRequestURL = os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	}
	cfg.IDTokenRequestToken = os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	return nil
}

func (cfg *oidcConfig) validate() clierror.Error {
	if cfg.IDTokenRequestURL == "" {
		return clierror.New(
			"ID token request URL is required",
			"make sure you're running the command in Github Actions environment",
			"provide id-token-request-url flag or ACTIONS_ID_TOKEN_REQUEST_URL env variable",
		)
	}

	if cfg.IDTokenRequestToken == "" {
		return clierror.New(
			"ACTIONS_ID_TOKEN_REQUEST_TOKEN env variable is required",
			"make sure you're running the command in Github Actions environment",
		)
	}
	return nil
}

func runOIDC(cfg *oidcConfig) clierror.Error {
	// get Github token
	token, err := getGithubToken(cfg.IDTokenRequestURL, cfg.IDTokenRequestToken, cfg.audience)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get token"))
	}

	enrichedKubeconfig, err := createKubeconfig(cfg.caCertificate, cfg.clusterServer, token)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create kubeconfig"))
	}

	if cfg.output != "" {
		err = clientcmd.WriteToFile(*enrichedKubeconfig, cfg.output)
		println("Kubeconfig saved to: " + cfg.output)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to save kubeconfig to file"))
		}
	} else {
		message, err := clientcmd.Write(*enrichedKubeconfig)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to print kubeconfig"))
		}
		fmt.Println(string(message))
	}

	return nil
}

func getGithubToken(url, requestToken, audience string) (string, error) {
	// create http client

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
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
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("failed to get token from server: %s", response.Status)
	}

	tokenData := TokenData{}
	err = json.NewDecoder(response.Body).Decode(&tokenData)
	if err != nil {
		return "", err
	}
	return tokenData.Value, nil
}

func createKubeconfig(caCertificate, clusterServer, token string) (*api.Config, error) {
	certificate, err := base64.StdEncoding.DecodeString(caCertificate)
	if err != nil {
		return nil, err
	}

	config := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"cluster": {
				Server:                   clusterServer,
				CertificateAuthorityData: certificate,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"user": {
				Token: token,
			},
		},
		Contexts: map[string]*api.Context{
			"default": {
				Cluster:  "cluster",
				AuthInfo: "user",
			},
		},
		CurrentContext: "default",
		Extensions:     nil,
	}

	return config, nil
}
