package access

import (
	"fmt"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"strconv"
	"strings"
)

type accessConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name        string
	clusterrole string
	output      string
	namespace   string
	time        string
	permanent   bool
}

func NewAccessCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := accessConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "access",
		Short: "Enrich kubeconfig with access",
		Long:  "Enrich kubeconfig with Service Account based token and certificate",
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(runAccess(&cfg))
		},
	}

	cfg.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&cfg.name, "name", "", "Name of the Service Account to be created")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "Name of the cluster role to bind the Service Account to")
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the output kubeconfig file")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace to create the resources in")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "How long should the token be valid for, by default 1h (use h for hours and d for days)")
	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Should the token be valid indefinitely")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("clusterrole")

	return cmd
}

func runAccess(cfg *accessConfig) clierror.Error {
	// Create objects
	clierr := createObjects(cfg)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to create objects"))
	}
	enrichedKubeconfig, clierr := prepareKubeconfig(cfg)
	if clierr != nil {
		return clierr
	}

	if cfg.output != "" {
		err := clientcmd.WriteToFile(*enrichedKubeconfig, cfg.output)
		fmt.Println("Kubeconfig saved to: " + cfg.output)
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

func prepareKubeconfig(cfg *accessConfig) (*api.Config, clierror.Error) {
	currentCtx := cfg.KubeClient.ApiConfig().CurrentContext
	clusterName := cfg.KubeClient.ApiConfig().Contexts[currentCtx].Cluster
	token, duration, err := getServiceAccountToken(cfg)
	if err != nil {
		return nil, err
	}

	fmt.Println("Token will expire: " + duration)

	var certData []byte
	if cfg.permanent == true {
		secret, err := cfg.KubeClient.Static().CoreV1().Secrets(cfg.namespace).Get(cfg.Ctx, cfg.name, metav1.GetOptions{})
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("failed to get secret"))
		}
		token = string(secret.Data["token"])
		certData = secret.Data["ca.crt"]
	} else {
		certData = cfg.KubeClient.ApiConfig().Clusters[clusterName].CertificateAuthorityData
	}
	// Create a new kubeconfig
	kubeconfig := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   cfg.KubeClient.ApiConfig().Clusters[clusterName].Server,
				CertificateAuthorityData: certData, //cfg.KubeClient.ApiConfig().Clusters[clusterName].CertificateAuthorityData,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			cfg.name: {
				Token: token,
			},
		},
		Contexts: map[string]*api.Context{
			currentCtx: {
				Cluster:   clusterName,
				Namespace: cfg.namespace,
				AuthInfo:  cfg.name,
			},
		},
		CurrentContext: currentCtx,
		Extensions:     nil,
	}
	return kubeconfig, nil
}

func createObjects(cfg *accessConfig) clierror.Error {
	err := createServiceAccount(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Service Account"))
	}

	err = createClusterRoleBinding(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Cluster Role Binding"))
	}

	return nil
}

func createServiceAccount(cfg *accessConfig) error {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.name,
			Namespace: cfg.namespace,
		},
	}
	_, err := cfg.KubeClient.Static().CoreV1().ServiceAccounts(cfg.namespace).Create(cfg.Ctx, &sa, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func getServiceAccountToken(cfg *accessConfig) (string, string, clierror.Error) {

	if cfg.permanent == true {
		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.name,
				Namespace: cfg.namespace,
				Annotations: map[string]string{
					"kubernetes.io/service-account.name": cfg.name,
				},
			},
			Type: v1.SecretTypeServiceAccountToken,
		}

		_, err := cfg.KubeClient.Static().CoreV1().Secrets(cfg.namespace).Create(cfg.Ctx, &secret, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return "", "", clierror.Wrap(err, clierror.New("failed to create secret"))
		}

		return "", "never", nil
	} else {

		var seconds int64
		if strings.Contains(cfg.time, "h") {
			// remove the "h" from the string
			cfg.time = strings.TrimRight(cfg.time, "h")
			// convert the string to an int
			hours, err := strconv.Atoi(cfg.time)
			if err != nil {
				return "", "", clierror.Wrap(err, clierror.New("failed to convert time to seconds", "Make sure to use h for hours and d for days"))
			}
			// convert the hours to seconds
			seconds = int64(hours * 3600)
		}

		if strings.Contains(cfg.time, "d") {
			// remove the "d" from the string
			cfg.time = strings.TrimRight(cfg.time, "d")
			// convert the string to an int
			days, err := strconv.Atoi(cfg.time)
			if err != nil {
				return "", "", clierror.Wrap(err, clierror.New("failed to convert time to seconds", "Make sure to use h for hours and d for days"))
			}
			// convert the days to seconds
			seconds = int64(days * 86400)
		}

		if seconds == 0 {
			return "", "", clierror.New("failed to convert the token duration", "Make sure to use h for hours and d for days")
		}

		tokenRequest := authv1.TokenRequest{
			Spec: authv1.TokenRequestSpec{
				ExpirationSeconds: &seconds,
			},
		}

		tokenResponse, err := cfg.KubeClient.Static().CoreV1().ServiceAccounts(cfg.namespace).CreateToken(cfg.Ctx, cfg.name, &tokenRequest, metav1.CreateOptions{})
		if err != nil {
			return "", "", clierror.Wrap(err, clierror.New("failed to create token"))
		}
		return tokenResponse.Status.Token, tokenResponse.Status.ExpirationTimestamp.String(), nil
	}
}

func createClusterRoleBinding(cfg *accessConfig) error {
	cRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.clusterrole + "-binding",
			Namespace: cfg.namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      cfg.name,
				Namespace: cfg.namespace,
			}},

		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: cfg.clusterrole,
		},
	}
	_, err := cfg.KubeClient.Static().RbacV1().ClusterRoleBindings().Create(cfg.Ctx, &cRoleBinding, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
