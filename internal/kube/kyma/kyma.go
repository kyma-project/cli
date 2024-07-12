package kyma

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetDefaultKyma uses dynamic client to get the default Kyma CR from the kyma-system namespace and cast it to the Kyma structure
func GetDefaultKyma(ctx context.Context, client kube.Client) (*Kyma, error) {
	u, err := client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Get(ctx, "default", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	kyma := &Kyma{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)

	return kyma, err
}

// UpdateDefaultKyma uses dynamic client to update the default Kyma CR from the kyma-system namespace based on the Kyma CR from arguments
func UpdateDefaultKyma(ctx context.Context, client kube.Client, obj *Kyma) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Update(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})

	return err
}
