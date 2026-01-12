package resources

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NamespaceExists(ctx context.Context, client kube.Client, namespace string) error {
	if namespace == "" {
		return nil
	}

	_, err := client.Static().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("namespace '%s' does not exist", namespace)
		}
		return fmt.Errorf("failed to verify namespace existence")
	}

	return nil
}
