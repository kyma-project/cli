package kube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd/api"

	istio "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
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

// NewFromRestConfigWithTimeout returns a KymaKube given already existing rest.Config. It means that the access to a single cluster is given explicitly.
// The returned KymaKube has an empty (but not nil) kubeCfg property. This is on purpose: we will not use this property to switch between contexts or access any remote clusters, the access to the single cluster is already provided by the config parameter.
func NewFromRestConfigWithTimeout(config *rest.Config, t time.Duration) (KymaKube, error) {

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

	//An empty k8s.io/client-go/tools/clientcmd/api.Config object.
	kubeConfig := &api.Config{
		Preferences: api.Preferences{
			Extensions: map[string]runtime.Object{},
		},
		Clusters:   map[string]*api.Cluster{},
		AuthInfos:  map[string]*api.AuthInfo{},
		Contexts:   map[string]*api.Context{},
		Extensions: map[string]runtime.Object{},
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

func (c *client) Apply(manifest []byte) error {
	// Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(c.restCfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// parse manifest
	objChan, parseErr := parseManifest(manifest)

	for {
		select {
		case data, ok := <-objChan:
			if !ok {
				return nil
			}
			if err := c.applyManifest(data, mapper); err != nil {
				return err
			}
		case err, ok := <-parseErr:
			if !ok {
				return nil
			}
			if err == nil {
				continue
			}
			return err
		}
	}
}

// parseManifest can parse a multi-doc yaml manifest and send back each unstructured object through the channel
func parseManifest(data []byte) (<-chan []byte, <-chan error) {
	chanErr := make(chan error)
	chanBytes := make(chan []byte)
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))

	go func() {
		defer close(chanErr)
		defer close(chanBytes)

		for {
			buf, err := multidocReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				chanErr <- errors.Wrap(err, "failed to read yaml data")
				return
			}
			chanBytes <- buf
		}
	}()
	return chanBytes, chanErr
}

// applyAsync applies the given manifest with the given mapping.
func (c *client) applyManifest(manifest []byte, mapper *restmapper.DeferredDiscoveryRESTMapper) error {
	// Decode YAML manifest into unstructured.Unstructured
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj := &unstructured.Unstructured{}
	_, gvk, err := decode(manifest, nil, obj)
	if err != nil {
		return err
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = c.dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = c.dynamic.Resource(mapping.Resource)
	}

	// Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// Create or Update the object with SSA
	// types.ApplyPatchType indicates SSA.
	_, err = dr.Patch(context.Background(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "kyma",
	})

	return err
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

func (c *client) WaitDeploymentStatus(namespace, name string, cond appsv1.DeploymentConditionType, status corev1.ConditionStatus) error {
	watchFn := func() (bool, error) {
		d, err := c.static.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil && k8sErrors.IsNotFound(err) {
			return false, err
		}

		for _, c := range d.Status.Conditions {
			if c.Type == cond {
				return c.Status == status, nil
			}
		}
		return false, nil
	}

	return watch(name, watchFn, c.restCfg.Timeout)
}

func (c *client) WaitPodStatus(namespace, name string, status corev1.PodPhase) error {
	watchFn := func() (bool, error) {
		pod, err := c.Static().CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return false, err
		}

		if status == pod.Status.Phase {
			return true, nil
		}
		return false, nil
	}

	return watch(name, watchFn, c.restCfg.Timeout)
}

func (c *client) WaitPodStatusByLabel(namespace, labelName, labelValue string, status corev1.PodPhase) error {
	watchFn := func() (bool, error) {
		pods, err := c.Static().CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)})
		if err != nil {
			return false, err
		}

		ok := true
		for _, pod := range pods.Items {
			// if any pod is not in the desired status no need to check further
			if status != pod.Status.Phase {
				ok = false
				break
			}
		}
		return ok, nil
	}

	return watch(fmt.Sprintf("Pod labeled: %s=%s", labelName, labelValue), watchFn, c.restCfg.Timeout)
}

func (c *client) WatchResource(res schema.GroupVersionResource, name, namespace string, checkFn func(u *unstructured.Unstructured) (bool, error)) error {
	watchFn := func() (bool, error) {
		var itm *unstructured.Unstructured
		var err error
		if namespace != "" {
			itm, err = c.Dynamic().Resource(res).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
		} else {
			itm, err = c.Dynamic().Resource(res).Get(context.Background(), name, metav1.GetOptions{})
		}
		if err != nil {
			return false, errors.Wrapf(err, "Failed to check %s", res.Resource)
		}
		return checkFn(itm)
	}

	return watch(res.Resource, watchFn, c.restCfg.Timeout)
}

func (c *client) DefaultNamespace() string {
	if ctx, ok := c.KubeConfig().Contexts[c.KubeConfig().CurrentContext]; ok && ctx.Namespace != "" {
		return ctx.Namespace
	}
	return defaultNamespace
}

// watch provides a unified implementation to watch resources.
// timeout of zero does NOT mean "no timeout", i.e. "wait forever" - it means the function will time-out almost immediately - at the discretion of the golang scheduler.
func watch(res string, watchFn func() (bool, error), timeout time.Duration) error {
	timeChan := time.After(timeout)

	for {
		select {
		case <-timeChan:
			return fmt.Errorf("Timeout reached while waiting for %s", res)

		default:
			finished, err := watchFn()
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
