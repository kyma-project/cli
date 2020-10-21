package hosts

import (
	"context"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/minikube"
	"github.com/kyma-project/cli/pkg/installation"
	"github.com/kyma-project/cli/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddDevDomainsToEtcHosts(
	s step.Step, clusterInfo installation.ClusterInfo, kymaKube kube.KymaKube, verbose bool, timeout time.Duration, domain string) error {
	hostnames := ""

	vsList, err := kymaKube.Istio().NetworkingV1alpha3().VirtualServices("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, v := range vsList.Items {
		for _, host := range v.Spec.Hosts {
			hostnames = hostnames + " " + host
		}
	}

	hostAlias := "127.0.0.1" + hostnames

	if clusterInfo.LocalVMDriver != "none" {
		_, err := minikube.RunCmd(verbose, clusterInfo.Profile, timeout, "ssh", "sudo /bin/sh -c 'echo \""+hostAlias+"\" >> /etc/hosts'")
		if err != nil {
			return err
		}
	}

	hostAlias = strings.Trim(clusterInfo.LocalIP, "\n") + hostnames

	return addDevDomainsToEtcHostsOSSpecific(domain, s, hostAlias)
}
