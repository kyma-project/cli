package access

import (
	"fmt"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type accessConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name        string
	clusterrole string
	kubeconfig  string
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
		Long:  "xxxxxx",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return cfg.KubeClientConfig.Complete()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAccess(&cfg)
		},
	}

	cfg.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&cfg.name, "name", "", "Name of the Service Account to be created")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "Name of the cluster role to bind the Service Account")
	//cmd.Flags().StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	cmd.Flags().StringVar(&cfg.output, "output", "???", "Path to the output kubeconfig file")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace ")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("clusterrole")

	return cmd
}

func runAccess(cfg *accessConfig) error {
	// Create objects
	err := createObjects(cfg)
	if err != nil {
		fmt.Sprintf("Error creating objects: %v", err)
		return err
	}

	return err
}

func createObjects(cfg *accessConfig) error {

	// Create ServiceAccount
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.name,
			Namespace: cfg.namespace,
		},
	}
	unstructuredSA, err := kube.ToUnstructured(sa, v1.SchemeGroupVersion.WithKind("ServiceAccount"))
	if err != nil {
		return err
	}
	_, err = cfg.KubeClient.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "ServiceAccount",
	}).Namespace("default").Create(cfg.Ctx, unstructuredSA, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create Secret
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
	unstructuredSecret, err := kube.ToUnstructured(secret, v1.SchemeGroupVersion.WithKind("Secret"))
	if err != nil {
		return err
	}
	_, err = cfg.KubeClient.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "Secret",
	}).Namespace("default").Create(cfg.Ctx, unstructuredSecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create ClusterRole
	cRole := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.clusterrole,
			Namespace: cfg.namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "batch", "apps", "gateway.kyma-project.io", "servicecatalog.k8s.io"},
				Resources: []string{"deployments", "replicaset", "pods", "jobs", "configmaps", "apirules", "serviceinstances", "servicebindings", "services", "secrets"},
				Verbs:     []string{"create", "update", "patch", "delete", "get", "list"},
			},
		},
	}
	unstructuredCRole, err := kube.ToUnstructured(cRole, rbacv1.SchemeGroupVersion.WithKind("ClusterRole"))
	if err != nil {
		return err
	}
	_, err = cfg.KubeClient.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "ClusterRole",
	}).Namespace("default").Create(cfg.Ctx, unstructuredCRole, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create ClusterRoleBinding
	cRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.name,
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
	unstructuredCRoleBinding, err := kube.ToUnstructured(cRoleBinding, rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"))
	if err != nil {
		return err
	}
	_, err = cfg.KubeClient.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "ClusterRoleBinding",
	}).Namespace("default").Create(cfg.Ctx, unstructuredCRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
