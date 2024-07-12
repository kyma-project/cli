package kyma

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const (
	defaultKymaName      = "default"
	defaultKymaNamespace = "kyma-system"
)

type Interface interface {
	GetDefaultKyma(context.Context) (*Kyma, error)
	UpdateDefaultKyma(context.Context, *Kyma) error
}

type client struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &client{
		dynamic: dynamic,
	}
}

// GetDefaultKyma uses dynamic client to get the default Kyma CR from the kyma-system namespace and cast it to the Kyma structure
func (c *client) GetDefaultKyma(ctx context.Context) (*Kyma, error) {
	u, err := c.dynamic.Resource(GVRKyma).
		Namespace(defaultKymaNamespace).
		Get(ctx, defaultKymaName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	kyma := &Kyma{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)

	return kyma, err
}

// UpdateDefaultKyma uses dynamic client to update the default Kyma CR from the kyma-system namespace based on the Kyma CR from arguments
func (c *client) UpdateDefaultKyma(ctx context.Context, obj *Kyma) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = c.dynamic.Resource(GVRKyma).
		Namespace(defaultKymaNamespace).
		Update(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})

	return err
}
