package kube

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli/pkg/api/octopus"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultWaitSleep   = 3 * time.Second
)

// client is the default KymaKube implementation
type client struct {
	static  kubernetes.Interface
	dynamic dynamic.Interface
	octps   octopus.OctopusInterface
}

// NewFromConfig creates a new Kubernetes client based on the given Kubeconfig either provided by URL (in-cluster config) or via file (out-of-cluster config).
func NewFromConfig(url, file string) (KymaKube, error) {
	return NewFromConfigWithTimeout(url, file, defaultHTTPTimeout)
}

// NewFromConfigWithTimeout creates a new Kubernetes client based on the given Kubeconfig either provided by URL (in-cluster config) or via file (out-of-cluster config).
// Allows to set a custom timeout for the Kubernetes HTTP client.
func NewFromConfigWithTimeout(url, file string, t time.Duration) (KymaKube, error) {

	config, err := clientcmd.BuildConfigFromFlags(url, file)
	if err != nil {
		return nil, err
	}

	config.Timeout = t

	sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	octClient, err := octopus.NewOctopusRESTClient(2 * time.Second)
	if err != nil {
		return nil, err
	}

	return &client{
			static:  sClient,
			dynamic: dClient,
			octps:   octClient,
		},
		nil

}

func (c *client) Static() kubernetes.Interface {
	return c.static
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *client) Octopus() octopus.OctopusInterface {
	return c.octps
}

func (c *client) IsPodDeployed(namespace, name string) (bool, error) {
	_, err := c.Static().CoreV1().Pods(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		} else {
			// actual errors
			return false, err
		}
	}
	return true, nil
}

func (c *client) IsPodDeployedByLabel(namespace, labelName, labelValue string) (bool, error) {
	pods, err := c.Static().CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)})
	if err != nil {
		return false, err
	}

	return len(pods.Items) > 0, nil
}

func (c *client) WaitPodStatus(namespace, name string, status corev1.PodPhase) error {
	for {
		pod, err := c.Static().CoreV1().Pods(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}

		if status == pod.Status.Phase {
			return nil
		}
		time.Sleep(defaultWaitSleep)
	}
}

func (c *client) WaitPodStatusByLabel(namespace, labelName, labelValue string, status corev1.PodPhase) error {
	for {
		pods, err := c.Static().CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)})
		if err != nil {
			return err
		}

		ok := true
		for _, pod := range pods.Items {
			// if any pod is not in the desired status no need to check further
			if status != pod.Status.Phase {
				ok = false
				break
			}
		}
		if ok {
			return nil
		}
		time.Sleep(defaultWaitSleep)
	}
}

// TODO we do not need more wait functions once deleteion is not done via Kubectl, the K8s API will wait on its own
func (c *client) WaitPodsGone(namespace, labelName, labelValue string) error {
	for {
		deployed, err := c.IsPodDeployedByLabel(namespace, labelName, labelValue)
		if err != nil {
			return err
		}

		if !deployed {
			return nil
		}
		time.Sleep(defaultWaitSleep)
	}
}
