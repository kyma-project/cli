package istio

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/clierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/client-go/dynamic"
)

const (
	GatewayName      = "kyma-gateway"
	GatewayNamespace = "kyma-system"
)

type Interface interface {
	GetClusterAddressFromGateway(ctx context.Context) (string, clierror.Error)
}

type client struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &client{
		dynamic: dynamic,
	}
}

func (c *client) GetClusterAddressFromGateway(ctx context.Context) (string, clierror.Error) {
	gvr := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
		Resource: "gateways",
	}
	u, err := c.dynamic.Resource(gvr).Namespace(GatewayNamespace).Get(ctx, GatewayName, metav1.GetOptions{})
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("while getting Gateway %s in namespace %s", GatewayName, GatewayNamespace))
	}
	var gateway v1beta1.Gateway
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &gateway)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("Failed to convert unstructured object to Gateway"))
	}

	// kyma gateway can't be modified by the user so we can assume that it has at least one server and host
	// this `if` should protect us when this situation changes
	if len(gateway.Spec.Servers) < 1 || len(gateway.Spec.Servers[0].Hosts) < 1 {
		return "", clierror.New(fmt.Sprintf("the Gateway %s in namespace %s does not have any hosts defined", GatewayName, GatewayNamespace))
	}

	host := gateway.Spec.Servers[0].Hosts[0]

	// host is always in format '*.<address>' so we need to remove the first two characters
	return host[2:], nil
}
