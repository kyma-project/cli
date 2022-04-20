package hosts

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	localDevDomain = "local.kyma.dev"
)

func GetVirtualServiceHostnames(kymaKube kube.KymaKube) ([]string, error) {
	vsList, err := kymaKube.Istio().NetworkingV1alpha3().VirtualServices("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hostnames := []string{}
	for _, v := range vsList.Items {
		hostnames = append(hostnames, v.Spec.Hosts...)
	}

	return hostnames, nil
}

func AddDevDomainsToEtcHostsKyma2(s step.Step, kymaKube kube.KymaKube) error {
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

	var gatewayAddress string
	if strings.Contains(hostnames, localDevDomain) {
		gatewayAddress = "127.0.0.1"
	} else {
		gatewayAddress, err = getIngressGatewayAddress(kymaKube)
		if err != nil {
			return err
		}
	}
	hostAlias := gatewayAddress + hostnames

	existingEntry, err := hasHostsEntry(hostAlias)
	if err != nil {
		return err
	}
	if existingEntry {
		s.LogInfof("Domain mappings are already existing in your 'hosts' file")
		return nil
	}

	return writeHostsEntry(hostAlias)
}

func getIngressGatewayAddress(kymaKube kube.KymaKube) (string, error) {
	svc, err := kymaKube.Static().CoreV1().Services("istio-system").Get(context.Background(), "istio-ingressgateway", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

func hasHostsEntry(hostAlias string) (bool, error) {
	f, err := os.Open(hostsFile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		if fileScanner.Text() == hostAlias {
			return true, nil
		}
	}
	return false, nil
}

func writeHostsEntry(hostAlias string) error {
	f, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(hostAlias + "\n")
	return err
}
