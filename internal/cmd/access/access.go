package access

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type accessConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name        string
	clusterrole string
	output      string
	namespace   string
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
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "Name of the cluster role to bind the Service Account")
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the output kubeconfig file")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace to create the resources in")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("clusterrole")

	return cmd
}

func runAccess(cfg *accessConfig) clierror.Error {
	// Create objects
	err := createObjects(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.Message("failed to create objects"))
	}
	enrichedKubeconfig, err := prepareKubeconfig(cfg)
	if err != nil {
		return clierror.Wrap(err, clierror.Message("failed to prepare kubeconfig"))
	}

	if cfg.output != "" {
		err = clientcmd.WriteToFile(*enrichedKubeconfig, cfg.output)
		println("Kubeconfig saved to: " + cfg.output)
		if err != nil {
			return clierror.Wrap(err, clierror.Message("failed to save kubeconfig to file"))
		}
	} else {
		message, err := clientcmd.Write(*enrichedKubeconfig)
		if err != nil {
			return clierror.Wrap(err, clierror.Message("failed to print kubeconfig"))
		}
		fmt.Println(string(message))

	}

	return nil
}

func prepareKubeconfig(cfg *accessConfig) (*api.Config, error) {
	secret, err := cfg.KubeClient.Static().CoreV1().Secrets(cfg.namespace).Get(cfg.Ctx, cfg.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	currentCtx := cfg.KubeClient.ApiConfig().CurrentContext
	clusterName := cfg.KubeClient.ApiConfig().Contexts[currentCtx].Cluster

	// Create a new kubeconfig
	kubeconfig := &api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   cfg.KubeClient.ApiConfig().Clusters[clusterName].Server,
				CertificateAuthorityData: secret.Data["ca.crt"],
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			cfg.name: {
				Token: string(secret.Data["token"]),
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

func createObjects(cfg *accessConfig) error {
	err := createServiceAccount(cfg)
	if err != nil {
		return err
	}

	err = createSecret(cfg)
	if err != nil {
		return err
	}

	err = createClusterRole(cfg)
	if err != nil {
		return err
	}

	err = createClusterRoleBinding(cfg)
	if err != nil {
		return err
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
	if errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func createSecret(cfg *accessConfig) error {
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
	if errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func createClusterRole(cfg *accessConfig) error {
	cRole := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.clusterrole,
			Namespace: cfg.namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "batch", "apps", "gateway.kyma-project.io", "servicecatalog.k8s.io"},
				Resources: []string{"deployments", "replicasets", "pods", "jobs", "configmaps", "apirules", "serviceinstances", "servicebindings", "services", "secrets"},
				Verbs:     []string{"create", "update", "patch", "delete", "get", "list"},
			},
		},
	}
	_, err := cfg.KubeClient.Static().RbacV1().ClusterRoles().Create(cfg.Ctx, &cRole, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return err
	}
	return nil
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
	if errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
