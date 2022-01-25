package kube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"

	istio "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultWaitSleep   = 3 * time.Second
	defaultNamespace   = "default"
)

// client is the default KymaKube implementation
type client struct {
	static  kubernetes.Interface
	dynamic dynamic.Interface
	istio   istio.Interface
	restCfg *rest.Config
	kubeCfg *api.Config
}

// NewFromConfig creates a new Kubernetes client based on the given Kubeconfig either provided by URL (in-cluster config) or via file (out-of-cluster config).
func NewFromConfig(url, file string) (KymaKube, error) {
	return NewFromConfigWithTimeout(url, file, defaultHTTPTimeout)
}

// NewFromConfigWithTimeout creates a new Kubernetes client based on the given Kubeconfig either provided by URL (in-cluster config) or via file (out-of-cluster config).
// Allows to set a custom timeout for the Kubernetes HTTP client.
func NewFromConfigWithTimeout(url, file string, t time.Duration) (KymaKube, error) {
	config, err := restConfig(url, file)
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

	istioClient, err := istio.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := kubeConfig(file)
	if err != nil {
		return nil, err
	}

	return &client{
			static:  sClient,
			dynamic: dClient,
			istio:   istioClient,
			restCfg: config,
			kubeCfg: kubeConfig,
		},
		nil
}

func (c *client) Static() kubernetes.Interface {
	return c.static
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *client) Istio() istio.Interface {
	return c.istio
}

func (c *client) RestConfig() *rest.Config {
	return c.restCfg
}

func (c *client) KubeConfig() *api.Config {
	return c.kubeCfg
}

func (c *client) IsPodDeployed(namespace, name string) (bool, error) {
	_, err := c.Static().CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		// actual errors
		return false, err
	}
	return true, nil
}

func (c *client) IsPodDeployedByLabel(namespace, labelName, labelValue string) (bool, error) {
	pods, err := c.Static().CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)})
	if err != nil {
		return false, err
	}

	return len(pods.Items) > 0, nil
}

func (c *client) WaitPodStatus(namespace, name string, status corev1.PodPhase) error {
	for {
		pod, err := c.Static().CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
		pods, err := c.Static().CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)})
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

func (c *client) WatchResource(res schema.GroupVersionResource, name, namespace string, checkFn func(u *unstructured.Unstructured) (bool, error)) error {
	var timeout <-chan time.Time
	if c.restCfg.Timeout > 0 {
		timeout = time.After(c.restCfg.Timeout)
	}

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout reached while waiting for %s", res.Resource)

		default:
			var itm *unstructured.Unstructured
			var err error
			if namespace != "" {
				itm, err = c.Dynamic().Resource(res).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
			} else {
				itm, err = c.Dynamic().Resource(res).Get(context.Background(), name, metav1.GetOptions{})
			}
			if err != nil {
				return errors.Wrapf(err, "Failed to check %s", res.Resource)
			}

			finished, err := checkFn(itm)
			if err != nil {
				return err
			}
			if finished {
				return nil
			}
			time.Sleep(defaultWaitSleep)
		}
	}
}

func (c *client) DefaultNamespace() string {
	if ctx, ok := c.KubeConfig().Contexts[c.KubeConfig().CurrentContext]; ok && ctx.Namespace != "" {
		return ctx.Namespace
	}
	return defaultNamespace
}
