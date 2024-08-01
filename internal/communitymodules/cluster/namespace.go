package cluster

import (
	"context"
	"github.com/kyma-project/cli.v3/internal/clierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AssureNamespace(ctx context.Context, client kubernetes.Interface, namespace string) clierror.Error {
	_, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return nil
	}
	_, err = client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return clierror.New("failed to create namespace")
	}
	return nil
}
