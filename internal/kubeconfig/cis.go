package kubeconfig

import (
	"strings"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/cis"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func GetFromCIS(credentialsPath string) (*api.Config, clierror.Error) {
	credentials, clierr := auth.LoadCISCredentials(credentialsPath)
	if clierr != nil {
		return nil, clierr
	}
	token, clierr := auth.GetOAuthToken(
		credentials.GrantType,
		credentials.UAA.URL,
		credentials.UAA.ClientID,
		credentials.UAA.ClientSecret,
	)
	if clierr != nil {
		var hints []string
		if strings.Contains(clierr.String(), "Internal Server Error") {
			hints = append(hints, "check if CIS grant type is set to client credentials")
		}

		return nil, clierror.WrapE(clierr, clierror.New("failed to get access token", hints...))
	}

	localCISClient := cis.NewLocalClient(credentials, token)
	kubeconfigString, clierr := localCISClient.GetKymaKubeconfig()
	if clierr != nil {
		return nil, clierror.WrapE(clierr, clierror.New("failed to get kubeconfig"))
	}

	kubeconfig, err := clientcmd.Load([]byte(kubeconfigString))
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to parse kubeconfig"))
	}
	return kubeconfig, nil
}
