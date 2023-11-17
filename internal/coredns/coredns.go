package coredns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/kyma-project/cli/internal/clusterinfo"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

const (
	coreDNSPatchTemplate = `
.:53 {
    errors
    health
    rewrite name regex {{ .DomainName}} istio-ingressgateway.istio-system.svc.cluster.local
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      pods insecure
      fallthrough in-addr.arpa ip6.arpa
    }
    hosts /etc/coredns/NodeHosts {
      reload 1s
      fallthrough
    }
    prometheus :9153
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}
`
	defaultHostsTemplate = `
    {{ .K3dRegistryIP}} {{ .K3dRegistryHost}}
`
	// Default domain names for coreDNS patch
	coreDNSLocalDomainName  = `(.*)\.local\.kyma\.dev`
	coreDNSRemoteDomainName = `(.*)\.kyma\.example\.com`
)

// Patch patches the CoreDNS configuration based on the overrides and the cloud provider.
func Patch(logger *zap.Logger, kubeClient kubernetes.Interface, hasCustomDomain bool, clusterInfo clusterinfo.Info, customHostsTemplate string) (cm *v1.ConfigMap, err error) {
	err = retry.Do(func() error {
		_, err := kubeClient.AppsV1().Deployments("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
		if err != nil {
			if apierr.IsNotFound(err) {
				logger.Info("CoreDNS not found, skipping CoreDNS config patch")
				return nil
			}
			return err
		}

		// patches contain each key and value that needs to be patched in the coredns configmap data field.
		patches, err := generatePatches(hasCustomDomain, clusterInfo, customHostsTemplate)
		if err != nil {
			return err
		}
		if len(patches) != 0 {
			cm, err = doPatch(logger, kubeClient, patches)
			return err
		}
		return nil
	}, retry.Delay(2*time.Second), retry.Attempts(3), retry.DelayType(retry.FixedDelay))

	return
}

func doPatch(logger *zap.Logger, kubeClient kubernetes.Interface, patches map[string]string) (cm *v1.ConfigMap, err error) {
	configMaps := kubeClient.CoreV1().ConfigMaps("kube-system")
	coreDNSConfigMap, err := configMaps.Get(context.TODO(), "coredns", metav1.GetOptions{})
	exists := true
	if err != nil {
		if apierr.IsNotFound(err) {
			exists = false
		} else {
			return cm, err
		}
	}

	if exists {
		logger.Info("Patching CoreDNS config")
		return patchCoreDNSConfigMap(logger, configMaps, coreDNSConfigMap, patches)
	}

	logger.Info("Corefile not found, creating new CoreDNS config")
	return createCoreDNSConfigMap(logger, configMaps, patches)
}

func patchCoreDNSConfigMap(logger *zap.Logger, configMaps corev1.ConfigMapInterface, coreDNSConfigMap *v1.ConfigMap, patch map[string]string) (cm *v1.ConfigMap, err error) {
	for k, v := range patch {
		coreDNSConfigMap.Data[k] = v
	}
	jsontext, err := json.Marshal(coreDNSConfigMap)
	if err != nil {
		return
	}

	cm, err = configMaps.Patch(context.TODO(), "coredns", types.StrategicMergePatchType, jsontext, metav1.PatchOptions{})
	if err != nil {
		logger.Error("Could not patch CoreDNS config")
	}
	return
}

func createCoreDNSConfigMap(logger *zap.Logger, configMaps corev1.ConfigMapInterface, patch map[string]string) (cm *v1.ConfigMap, err error) {
	cm, err = configMaps.Create(context.TODO(), newCoreDNSConfigMap(patch), metav1.CreateOptions{})
	if err != nil {
		logger.Error("Could not create new CoreDNS config")
	}
	return
}

func newCoreDNSConfigMap(data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "coredns"},
		Data:       data,
	}
}

func generatePatches(hasCustomDomain bool, clusterInfo clusterinfo.Info, customHostsTemplate string) (map[string]string, error) {
	var err error
	patches := make(map[string]string)

	// patch the CoreFile only if not on gardener and no custom domain is provided
	if _, isGardener := clusterInfo.(clusterinfo.Gardener); !isGardener && !hasCustomDomain {
		var domainName string

		if _, isK3d := clusterInfo.(clusterinfo.K3d); isK3d {
			domainName = coreDNSLocalDomainName
		} else {
			domainName = coreDNSRemoteDomainName
		}
		patches["Corefile"], err = generateCorefile(domainName)
		if err != nil {
			return nil, err
		}
	}

	// Patch NodeHosts only on K3d
	if k3d, isK3d := clusterInfo.(clusterinfo.K3d); isK3d {
		patches["NodeHosts"], err = generateHosts(k3d.ClusterName, customHostsTemplate)
		if err != nil {
			return nil, err
		}
	}
	return patches, nil
}

func generateCorefile(domainName string) (coreFile string, err error) {
	patchvars := struct {
		DomainName string
	}{
		DomainName: domainName,
	}
	patchTemplate := template.Must(template.New("").Parse(coreDNSPatchTemplate))
	patchBuffer := new(bytes.Buffer)
	if err = patchTemplate.Execute(patchBuffer, patchvars); err != nil {
		return
	}

	coreFile = patchBuffer.String()
	return
}

func generateHosts(clusterName string, customHostsTemplate string) (string, error) {

	registryIP, err := k3dRegistryIP(clusterName)
	if err != nil {
		return "", err
	}
	hostsTemplate := defaultHostsTemplate
	if customHostsTemplate != "" {
		hostsTemplate = customHostsTemplate
	}
	patchVars := struct {
		K3dRegistryHost string
		K3dRegistryIP   string
	}{
		K3dRegistryHost: fmt.Sprintf("k3d-%s-registry", clusterName),
		K3dRegistryIP:   registryIP,
	}
	t := template.Must(template.New("").Parse(hostsTemplate))
	b := new(bytes.Buffer)
	if err := t.Execute(b, patchVars); err != nil {
		return "", err
	}

	return b.String(), nil
}

// the defaultInspector uses the standard docker client to get container information from the daemon in the local ENV
var defaultInspector = func(ctx context.Context, containerID string) (dockerTypes.ContainerJSON, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return dockerTypes.ContainerJSON{}, err
	}

	return client.ContainerInspect(context.Background(), containerID)
}

// registryIP uses docker inspect to find out the registry IP address in k3d
func k3dRegistryIP(cluster string) (string, error) {
	c, err := defaultInspector(context.Background(), fmt.Sprintf("k3d-%s-registry", cluster))
	if err != nil {
		return "", err
	}

	net, exists := c.NetworkSettings.Networks[fmt.Sprintf("k3d-%s", cluster)]
	if !exists {
		return "", errors.New("could not find network settings in k3d registry")
	}

	if net.IPAddress == "" {
		return "", errors.New("k3d registry IP is empty")
	}

	return net.IPAddress, nil
}
