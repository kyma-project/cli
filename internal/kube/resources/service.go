package resources

import (
	"context"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/istio"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func CreateService(ctx context.Context, client kube.Client, name, namespace string, port int32) error {
	service := buildService(name, namespace, port)
	_, err := client.Static().CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func CreateAPIRule(ctx context.Context, client rootlessdynamic.Interface, name, namespace, host string, port uint32) error {
	apirule := buildAPIRule(name, namespace, host, port)
	uAPIRule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apirule)
	if err != nil {
		return err
	}
	return client.Apply(ctx, &unstructured.Unstructured{Object: uAPIRule}, false)
}

func buildService(name, namespace string, port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt32(port),
				},
			},
		},
	}
}

func buildAPIRule(name, namespace, host string, port uint32) *v2alpha1.APIRule {
	return &v2alpha1.APIRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v2alpha1",
			Kind:       "APIRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: v2alpha1.APIRuleSpec{
			Hosts: []*v2alpha1.Host{
				ptr.To(v2alpha1.Host(host)),
			},
			Gateway: ptr.To(fmt.Sprintf("%s/%s", istio.DefaultGatewayNamespace, istio.DefaultGatewayName)),
			Rules: []v2alpha1.Rule{
				{
					Path:    "/*",
					Methods: []v2alpha1.HttpMethod{"GET", "POST", "PUT", "DELETE", "PATCH"},
					NoAuth:  ptr.To(true),
				},
			},
			Service: &v2alpha1.Service{
				Name:      ptr.To(name),
				Namespace: ptr.To(namespace),
				Port:      &port,
			},
		},
	}
}
