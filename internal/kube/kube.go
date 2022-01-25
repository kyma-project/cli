package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	istio "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

//go:generate mockery --name KymaKube

// KymaKube defines the Kyma-enhanced kubernetes API.
// It provides all functionality of the Kubernetes API plus extra functionality
type KymaKube interface {
	Static() kubernetes.Interface
	Dynamic() dynamic.Interface
	Istio() istio.Interface

	// RestConfig provides the REST configuration of the kubernetes client
	RestConfig() *rest.Config

	// KubeConfig provides the currently used kubeconfig
	KubeConfig() *api.Config

	// DefaultNamespace finds out what the default namespace is based on:
	// 1. Default namespace on the Kubeconfig
	// 2. Default cluster namespace constant
	DefaultNamespace() string

	// IsPodDeployed checks if a pod is in the given namespace (independently of its status)
	IsPodDeployed(namespace, name string) (bool, error)

	// IsPodDeployedByLabel checks if there is at least 1 pod in the given namespace with the given label  (independently of its status)
	IsPodDeployedByLabel(namespace, labelName, labelValue string) (bool, error)

	// WaitPodStatus waits for the given pod to reach the desired status.
	WaitPodStatus(namespace, name string, status corev1.PodPhase) error

	// WaitPodStatusByLabel selects a set of pods by label and waits for them
	WaitPodStatusByLabel(namespace, labelName, labelValue string, status corev1.PodPhase) error

	// WatchResource watches an arbitrary resource using the k8s unstructured API.
	// To check if the resource is in the desired state, checkFn is called repeatedly passing the resource as parameter,
	// until either it returns true or the timeout is reached.
	// If the timeout is reached an error is returned.
	WatchResource(res schema.GroupVersionResource, name, namespace string, checkFn func(u *unstructured.Unstructured) (bool, error)) error
}
