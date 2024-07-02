package kyma

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func GetDefaultKyma(ctx context.Context, client kube.Client) (*unstructured.Unstructured, error) {
	return client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Get(ctx, "default", metav1.GetOptions{})
}

func UpdateDefaultKyma(ctx context.Context, client kube.Client, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Update(ctx, obj, metav1.UpdateOptions{})
}
