package kube

import (
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

// KymaKube defines the Kyma-enhanced kubernetes API.
// It provides all functionality of the Kubernetes API plus extra functionality
type KymaKube interface {
	kubernetes.Interface

	// IsPodDeployed checks if a pod is in the given namespace (independently of its status)
	IsPodDeployed(namespace, name string) (bool, error)

	// IsPodDeployedByLabel checks if a pod is in the given namespace (independently of its status)
	IsPodDeployedByLabel(namespace, labelName, labelValue string) (bool, error)

	// WaitPodStatus waits for the given pod to rech the desired status.
	WaitPodStatus(namespace, name string, status corev1.PodPhase) error

	// WaitPodStatusByLabel selects a set of pods by label and waits for them
	WaitPodStatusByLabel(namespace, labelName, labelValue string, status corev1.PodPhase) error

	// TODO we do not need more wait functions once deleteion is not done via Kubectl, the K8s API will wait on its own
	WaitPodGone(namespace, name string) error

	// JSONPath allows to run a JSON query on a given Kubernetes resource
	JSONPath(res interface{}, query string) (string, error)
}
