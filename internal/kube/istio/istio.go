package istio

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/client-go/dynamic"
)

const (
	DefaultGatewayName      = "kyma-gateway"
	DefaultGatewayNamespace = "kyma-system"
)

type Interface interface {
	GetHostFromVirtualServiceByApiruleName(ctx context.Context, name, namespace string) (string, clierror.Error)
}

type client struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &client{
		dynamic: dynamic,
	}
}

func (c *client) GetHostFromVirtualServiceByApiruleName(ctx context.Context, name, namespace string) (string, clierror.Error) {
	//TODO: rework after https://github.com/kyma-project/api-gateway/issues/2264 is resolved

	//Wait for the virtual service to be created
	gvr := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1",
		Resource: "virtualservices",
	}
	u, err := c.dynamic.Resource(gvr).Namespace(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("apirule.gateway.kyma-project.io/v1beta1=%s.%s", name, namespace),
	})
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("while watching VirtualServices in namespace %s", namespace))
	}
	defer u.Stop()

	ch := u.ResultChan()
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return "", clierror.New("watch channel closed before VirtualService was created")
			}
			if event.Type == "ADDED" || event.Type == "MODIFIED" {
				obj := event.Object.(*unstructured.Unstructured)
				var vs v1beta1.VirtualService
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &vs)
				if err != nil {
					return "", clierror.Wrap(err, clierror.New("Failed to convert unstructured object to VirtualService"))
				}
				if len(vs.Spec.Hosts) < 1 {
					return "", clierror.New("the VirtualService does not have any hosts defined")
				}
				return vs.Spec.Hosts[0], nil
			}
		case <-ctx.Done():
			return "", clierror.New("timed out waiting for VirtualService to be created")
		}
	}
}
