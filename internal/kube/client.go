package kube

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	lifecycleManagerApi "github.com/kyma-project/lifecycle-manager/api"
	"github.com/pkg/errors"
	istio "istio.io/client-go/pkg/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd/api"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	v1extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultWaitSleep   = 3 * time.Second
	defaultNamespace   = "default"
)

// client is the default KymaKube implementation
type client struct {
	static     kubernetes.Interface
	dynamic    dynamic.Interface
	istio      istio.Interface
	restCfg    *rest.Config
	kubeCfg    *api.Config
	disco      discovery.DiscoveryInterface
	mapper     meta.ResettableRESTMapper
	ctrlClient ctrlClient.WithWatch
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

	disco, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))

	newScheme := scheme.Scheme

	if err := v1extensions.AddToScheme(newScheme); err != nil {
		return nil, err
	}
	if err := lifecycleManagerApi.AddToScheme(newScheme); err != nil {
		return nil, err
	}

	nClient, err := ctrlClient.NewWithWatch(config, ctrlClient.Options{Scheme: newScheme})
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
			static:     sClient,
			dynamic:    dClient,
			istio:      istioClient,
			restCfg:    config,
			kubeCfg:    kubeConfig,
			disco:      disco,
			mapper:     mapper,
			ctrlClient: nClient,
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

	disco, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))

	newScheme := scheme.Scheme

	if err := v1extensions.AddToScheme(newScheme); err != nil {
		return nil, err
	}
	if err := lifecycleManagerApi.AddToScheme(newScheme); err != nil {
		return nil, err
	}

	ctrlClient, err := ctrlClient.NewWithWatch(config, ctrlClient.Options{Scheme: newScheme})
	if err != nil {
		return nil, err
	}

	return &client{
			static:     sClient,
			dynamic:    dClient,
			istio:      istioClient,
			restCfg:    config,
			kubeCfg:    kubeConfig,
			disco:      disco,
			mapper:     mapper,
			ctrlClient: ctrlClient,
		},
		nil
}

func (c *client) Static() kubernetes.Interface {
	return c.static
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *client) Ctrl() ctrlClient.WithWatch {
	return c.ctrlClient
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

func (c *client) ParseManifest(manifest []byte) ([]ctrlClient.Object, error) {
	// parse manifest
	manifests, err := parseManifest(manifest)
	if err != nil {
		return nil, err
	}

	objs := make([]ctrlClient.Object, 0, len(manifests))
	for _, manifest := range manifests {

		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(manifest, obj); err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}

	return objs, nil
}

func (c *client) Apply(ctx context.Context, force bool, objs ...ctrlClient.Object) error {
	infos := make([]*resource.Info, 0, len(objs))
	for _, obj := range objs {
		gvk := obj.GetObjectKind().GroupVersionKind()
		mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if meta.IsNoMatchError(err) {
			c.mapper.Reset()
			mapping, _ = c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
			err = nil
		}
		if err != nil {
			return err
		}

		infos = append(
			infos, &resource.Info{
				Name:            obj.GetName(),
				Namespace:       obj.GetNamespace(),
				ResourceVersion: obj.GetResourceVersion(),
				Object:          obj,
				Mapping:         mapping,
			},
		)
	}
	return ConcurrentSSA(c.ctrlClient, "kyma", force).Run(ctx, infos)
}

// parseManifest can parse a multi-doc yaml manifest and send back each unstructured object through the channel
func parseManifest(data []byte) ([][]byte, error) {
	var chanBytes [][]byte
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))

	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				return chanBytes, nil
			}
			return nil, errors.Wrap(err, "failed to read yaml data")
		}
		chanBytes = append(chanBytes, buf)
	}
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
	pods, err := c.Static().CoreV1().Pods(namespace).List(
		context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)},
	)
	if err != nil {
		return false, err
	}

	return len(pods.Items) > 0, nil
}

func (c *client) WaitDeploymentStatus(
	namespace, name string, cond appsv1.DeploymentConditionType, status corev1.ConditionStatus,
) error {
	watchFn := func(ctx context.Context) (bool, error) {
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

	return watch(context.Background(), name, watchFn, c.restCfg.Timeout)
}

func (c *client) WaitPodStatus(namespace, name string, status corev1.PodPhase) error {
	watchFn := func(ctx context.Context) (bool, error) {
		pod, err := c.Static().CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return false, err
		}

		if status == pod.Status.Phase {
			return true, nil
		}
		return false, nil
	}

	return watch(context.Background(), name, watchFn, c.restCfg.Timeout)
}

func (c *client) WaitPodStatusByLabel(namespace, labelName, labelValue string, status corev1.PodPhase) error {
	watchFn := func(ctx context.Context) (bool, error) {
		pods, err := c.Static().CoreV1().Pods(namespace).List(
			ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, labelValue)},
		)
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

	return watch(
		context.Background(), fmt.Sprintf("Pod labeled: %s=%s", labelName, labelValue), watchFn, c.restCfg.Timeout,
	)
}

func (c *client) WatchResource(
	res schema.GroupVersionResource, name, namespace string, checkFn func(u *unstructured.Unstructured) (bool, error),
) error {
	watchFn := func(ctx context.Context) (bool, error) {
		var itm *unstructured.Unstructured
		var err error
		if namespace != "" {
			itm, err = c.Dynamic().Resource(res).Namespace(namespace).Get(
				ctx, name, metav1.GetOptions{},
			)
		} else {
			itm, err = c.Dynamic().Resource(res).Get(ctx, name, metav1.GetOptions{})
		}
		if err != nil {
			return false, errors.Wrapf(err, "Failed to check %s", res.Resource)
		}
		return checkFn(itm)
	}

	return watch(context.Background(), res.Resource, watchFn, c.restCfg.Timeout)
}

func (c *client) WatchObject(
	ctx context.Context, obj ctrlClient.Object, checkFn func(u ctrlClient.Object) (bool, error),
) error {

	key := ctrlClient.ObjectKeyFromObject(obj)
	watchFn := func(ctx context.Context) (bool, error) {
		if err := c.ctrlClient.Get(ctx, key, obj); err != nil {
			return false, errors.Wrapf(err, "Failed to check %s", key)
		}
		return checkFn(obj)
	}

	return watch(ctx, key.String(), watchFn, c.restCfg.Timeout)
}

func (c *client) DefaultNamespace() string {
	if ctx, ok := c.KubeConfig().Contexts[c.KubeConfig().CurrentContext]; ok && ctx.Namespace != "" {
		return ctx.Namespace
	}
	return defaultNamespace
}

// watch provides a unified implementation to watch resources.
// timeout of zero means "no timeout", i.e. "wait forever"
func watch(
	ctx context.Context, res string, watchFn func(ctx context.Context) (bool, error), timeout time.Duration,
) error {
	if timeout.Milliseconds() != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		finished, err := watchFn(ctx)
		if err != nil {
			return fmt.Errorf("error waiting for %s: %w", res, err)
		}
		if finished {
			return nil
		}
		time.Sleep(defaultWaitSleep)
	}
}
