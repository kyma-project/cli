package kubeconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

func PrepareWithToken(apiBase *api.Config, token string) *api.Config {
	currentUser := apiBase.Contexts[apiBase.CurrentContext].AuthInfo
	config := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   apiBase.Clusters,
		AuthInfos: map[string]*api.AuthInfo{
			currentUser: {
				Token: token,
			},
		},
		Contexts:       apiBase.Contexts,
		CurrentContext: apiBase.CurrentContext,
		Extensions:     apiBase.Extensions,
		Preferences:    apiBase.Preferences,
	}

	return config
}

func Prepare(ctx context.Context, client kube.Client, name, namespace, time, output string, permanent bool) (*api.Config, clierror.Error) {
	clusterName := getKubeconfigCurrentClusterName(client.APIConfig())
	var tokenData authv1.TokenRequestStatus
	var certData []byte
	var err clierror.Error

	// Prepare the token and certificate data
	if permanent {
		var secret *v1.Secret
		for ok := true; ok; ok = string(secret.Data["token"]) == "" {
			var loopErr error
			secret, loopErr = client.Static().CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
			if loopErr != nil {
				return nil, clierror.Wrap(loopErr, clierror.New("failed to get secret"))
			}
		}

		tokenData.Token = string(secret.Data["token"])
		certData = secret.Data["ca.crt"]
		if output != "" {
			fmt.Println("Token is valid permanently")
		}
	} else {
		certData = client.APIConfig().Clusters[clusterName].CertificateAuthorityData
		tokenData, err = getServiceAccountToken(ctx, client, name, namespace, time)
		if err != nil {
			return nil, err
		}
		if output != "" {
			fmt.Println("Token will expire: " + tokenData.ExpirationTimestamp.String())
		}
	}

	// Create a new kubeconfig
	kubeconfig := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   client.APIConfig().Clusters[clusterName].Server,
				CertificateAuthorityData: certData,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			name: {
				Token: tokenData.Token,
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:   clusterName,
				Namespace: namespace,
				AuthInfo:  name,
			},
		},
		CurrentContext: clusterName,
		Extensions:     nil,
	}

	return kubeconfig, nil
}

func PrepareFromOpenIDConnectorResource(ctx context.Context, client kube.Client, name string) (*api.Config, clierror.Error) {
	clusterName := getKubeconfigCurrentClusterName(client.APIConfig())

	oidcResUnstruct, err := client.Dynamic().Resource(OpenIdConnectGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New(fmt.Sprintf("failed to get %s oidc resource", name)))
	}

	var oidc OpenIDConnect
	if oidc, err = unmarshalOIDCResource(oidcResUnstruct.Object); err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to unmarshal OIDC resource"))
	}

	kubeconfig := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   client.APIConfig().Clusters[clusterName].Server,
				CertificateAuthorityData: client.APIConfig().Clusters[clusterName].CertificateAuthorityData,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			name: {
				Exec: &api.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1beta1",
					Command:    "kubectl-oidc_login",
					Args: []string{
						"get-token",
						"--oidc-issuer-url=" + oidc.Spec.IssuerURL,
						"--oidc-client-id=" + oidc.Spec.ClientID,
						"--oidc-extra-scope=email",
						"--oidc-extra-scope=openid",
					},
				},
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: name,
			},
		},
		CurrentContext: clusterName,
		Extensions:     nil,
	}

	return kubeconfig, nil
}

func getKubeconfigCurrentClusterName(kubeconfig *api.Config) string {
	clusterName := kubeconfig.CurrentContext
	if clusterName != "" {
		// current context exists
		return kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	}

	// get first cluster from map
	for clusterName = range kubeconfig.Clusters {
		break
	}

	return clusterName
}

func unmarshalOIDCResource(obj map[string]any) (OpenIDConnect, error) {
	var oidc OpenIDConnect
	b, err := json.Marshal(obj)
	if err != nil {
		return oidc, err
	}
	if err := json.Unmarshal(b, &oidc); err != nil {
		return oidc, err
	}
	return oidc, nil
}

func getServiceAccountToken(ctx context.Context, client kube.Client, name, namespace, time string) (authv1.TokenRequestStatus, clierror.Error) {
	var tokenData authv1.TokenRequestStatus

	seconds, errTime := parseExpirationTime(time)
	if errTime != nil {
		return tokenData, errTime
	}

	tokenRequest := authv1.TokenRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: &seconds,
		},
	}

	tokenResponse, err := client.Static().CoreV1().ServiceAccounts(namespace).CreateToken(ctx, name, &tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return tokenData, clierror.Wrap(err, clierror.New("failed to create token"))
	}
	return tokenResponse.Status, nil
}

func parseExpirationTime(time string) (int64, clierror.Error) {
	var seconds int64

	// Convert the time passed in argument to seconds
	if strings.Contains(time, "h") {
		// remove the "h" from the string
		time = strings.TrimRight(time, "h")
		// convert the string to an int
		hours, err := strconv.Atoi(time)
		if err != nil {
			return 0, clierror.Wrap(err, clierror.New("failed to convert time to seconds", "Make sure to use h for hours and d for days"))
		}
		// convert the hours to seconds
		seconds = int64(hours * 3600)
	}

	if strings.Contains(time, "d") {
		// remove the "d" from the string
		time = strings.TrimRight(time, "d")
		// convert the string to an int
		days, err := strconv.Atoi(time)
		if err != nil {
			return 0, clierror.Wrap(err, clierror.New("failed to convert time to seconds", "Make sure to use h for hours and d for days"))
		}
		// convert the days to seconds
		seconds = int64(days * 86400)
	}

	if seconds == 0 {
		return 0, clierror.New("failed to convert the token duration", "Make sure to use h for hours and d for days")
	}
	return seconds, nil
}
