package deploy

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// LocalKymaDevDomain contains the default domain name if Kyma is locally deployed
const LocalKymaDevDomain = "local.kyma.dev"

// KymaLocalDomain configures the local kyma domain in Kubernetes DNS
type KymaLocalDomain struct {
	coreDNSConfig []byte
	kubeClient    kubernetes.Interface
}

// New creates a new Kyma local domain instance
func newKymaLocalDomain(kubeClient kubernetes.Interface) *KymaLocalDomain {
	return &KymaLocalDomain{
		kubeClient: kubeClient,
		coreDNSConfig: []byte(`---
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
`),
	}
}

// ConfigureK8sDNS configure CoreDNS to support the local kyma domain
func (kld *KymaLocalDomain) ConfigureK8sDNS() error {

	_, err := kld.kubeClient.CoreV1().ConfigMaps("kube-system").Patch(
		context.Background(),
		"coredns",
		types.ApplyPatchType,
		kld.coreDNSConfig,
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
