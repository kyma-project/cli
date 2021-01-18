package deploy

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// LocalKymaDevDomain is the local Kyma domain name (used as default domain if Kyma is installed locally)
const LocalKymaDevDomain = "local.kyma.dev"

// DNSConfigurer modifies the Kubernetes DNS system to resolve *.local.kyma.dev names to locally running
// services. This is only required if Kyma runs locally (e.g. in a local K3s cluster) and uses the local Kyma dev domain.
//
// Per default, the DNS zone local.kyma.dev is managed by a public DNS system (which resolves names always to 127.0.0.1)
// to simplify the local setup of Kyma. But some sub-domains (e.g dex.local.kyma.dev) have to point to locally running
// Kubernetes services and have to return their internal IP address. This requires a patch of the Kubernetes DNS resolver.
type DNSConfigurer struct {
	kubeClient kubernetes.Interface
}

// NewDNSConfigurer creates a new DNS configurer instance
func NewDNSConfigurer(kubeClient kubernetes.Interface) *DNSConfigurer {
	return &DNSConfigurer{
		kubeClient: kubeClient,
	}
}

// ConfigureCoreDNS patches the Kubernetes CoreDNS to support the local kyma domain.
// It adds the rewrite plugin of CoreDNS to delegate requests to *.local.kyma.dev to the Istio Ingress Gateway.
func (kld *DNSConfigurer) ConfigureCoreDNS() error {
	coreDNSPatch := []byte(`---
apiVersion: v1
kind: ConfigMap
data:
  Corefile: |
    .:53 {
        errors
        health
        rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          upstream
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
`)

	_, err := kld.kubeClient.CoreV1().ConfigMaps("kube-system").Patch(
		context.Background(),
		"coredns",
		types.ApplyPatchType,
		coreDNSPatch,
		metav1.PatchOptions{
			FieldManager: "kyma-cli",
			Force: func() *bool { // use closure to get a boolean-pointer
				b := true
				return &b
			}(),
		},
	)

	return err
}
