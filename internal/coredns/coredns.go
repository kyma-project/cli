package coredns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-project/cli/internal/gardener"
	"github.com/kyma-project/cli/internal/overrides"
	"html/template"
	"time"

	"github.com/avast/retry-go"
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
	hostsTemplate = `
    {{ .K3sRegistryIP}} {{ .K3sRegistryHost}}
`
	// Default domain names for coreDNS patch
	coreDNSLocalDomainName  = `(.*)\.local\.kyma\.dev`
	coreDNSRemoteDomainName = `(.*)\.kyma\.example\.com`
)

// Patch patches the CoreDNS cnfiguration based on the overrides and the cloud provider.
func Patch(kubeClient kubernetes.Interface, overrides *overrides.Builder, isK3d bool) (cm *v1.ConfigMap, err error) {
	err = retry.Do(func() error {
		_, err := kubeClient.AppsV1().Deployments("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
		if err != nil {
			if apierr.IsNotFound(err) {
				//log.Info("CoreDNS not found, skipping CoreDNS config patch")
				return nil
			}
			return err
		}

		// patches contains each key and value that needs to be patched in the coredns configmap data field.
		patches, err := generatePatches(kubeClient, overrides, isK3d)
		if err != nil {
			return err
		}
		if len(patches) != 0 {
			cm, err = doPatch(kubeClient, patches)
			return err
		}
		return nil
	}, retry.Delay(2*time.Second), retry.Attempts(3), retry.DelayType(retry.FixedDelay))

	return
}

func doPatch(kubeClient kubernetes.Interface, patches map[string]string) (cm *v1.ConfigMap, err error) {
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
		//log.Info("Patching CoreDNS config")
		return patchCoreDNSConfigMap(configMaps, coreDNSConfigMap, patches)
	} else {
		//log.Info("Corefile not found, creating new CoreDNS config")
		return createCoreDNSConfigMap(configMaps, patches)
	}
}

func patchCoreDNSConfigMap(configMaps corev1.ConfigMapInterface, coreDNSConfigMap *v1.ConfigMap, patch map[string]string) (cm *v1.ConfigMap, err error) {
	for k, v := range patch {
		coreDNSConfigMap.Data[k] = v
	}
	jsontext, err := json.Marshal(coreDNSConfigMap)
	if err != nil {
		return
	}

	cm, err = configMaps.Patch(context.TODO(), "coredns", types.StrategicMergePatchType, jsontext, metav1.PatchOptions{})
	if err != nil {
		//log.Error("Could not patch CoreDNS config")
	}
	return
}

func createCoreDNSConfigMap(configMaps corev1.ConfigMapInterface, patch map[string]string) (cm *v1.ConfigMap, err error) {
	cm, err = configMaps.Create(context.TODO(), newCoreDNSConfigMap(patch), metav1.CreateOptions{})
	if err != nil {
		//log.Error("Could not create new CoreDNS config")
	}
	return
}

func newCoreDNSConfigMap(data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "coredns"},
		Data:       data,
	}
}

func generatePatches(kubeClient kubernetes.Interface, overrides *overrides.Builder, isK3s bool) (map[string]string, error) {
	patches := make(map[string]string)
	// patch the CoreFile only if not on gardener and no custom domain is provided
	gardenerDomain, err := gardener.Domain(kubeClient)
	if err != nil {
		return nil, err
	}
	o, err := overrides.Raw()
	if err != nil {
		return nil, err
	}
	_, hasCustomDomain := o.Find("global.domainName")
	if gardenerDomain == "" && !hasCustomDomain {
		var domainName string
		if isK3s {
			domainName = coreDNSLocalDomainName
		} else {
			domainName = coreDNSRemoteDomainName
		}
		patches["Corefile"], err = generateCorefile(domainName)
		if err != nil {
			return nil, err
		}
	}

	// Patch NodeHosts only on K3s
	if isK3s {
		patches["NodeHosts"], err = generateHosts(kubeClient)
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

func generateHosts(kubeClient kubernetes.Interface) (string, error) {
	clusterName, err := overrides.K3dClusterName(kubeClient)
	if err != nil {
		return "", err
	}
	registryIP, err := k3sRegistryIP(clusterName)
	if err != nil {
		return "", err
	}

	patchVars := struct {
		K3sRegistryHost string
		K3sRegistryIP   string
	}{
		K3sRegistryHost: fmt.Sprintf("k3d-%s-registry", clusterName),
		K3sRegistryIP:   registryIP,
	}
	t := template.Must(template.New("").Parse(hostsTemplate))
	b := new(bytes.Buffer)
	if err := t.Execute(b, patchVars); err != nil {
		return "", err
	}

	return b.String(), nil

}

// Abstract the docker container inspect to be able to test the k3s coreDNS patching
type containerInspector func(ctx context.Context, containerID string) (dockerTypes.ContainerJSON, error)

// the defaultInspector uses the standard docker client to get container information from the daemon in the local ENV
var defaultInspector = func(ctx context.Context, containerID string) (dockerTypes.ContainerJSON, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return dockerTypes.ContainerJSON{}, err
	}

	return client.ContainerInspect(context.Background(), containerID)
}

// registryIP uses docker inspect to find out the registry IP address in k3s
func k3sRegistryIP(cluster string) (string, error) {
	c, err := defaultInspector(context.Background(), fmt.Sprintf("k3d-%s-registry", cluster))
	if err != nil {
		return "", err
	}

	net, exists := c.NetworkSettings.Networks[fmt.Sprintf("k3d-%s", cluster)]
	if !exists {
		return "", errors.New("Could not find network settings in k3s registry.")
	}

	if net.IPAddress == "" {
		return "", errors.New("K3s registry IP is empty")
	}

	return net.IPAddress, nil
}
