package hostsfile

import (
	"context"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/txn2/txeh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//LocalIP contains the lLocal IP address
	LocalIP string = "127.0.0.1"
)

// Hostsfile handles local host file (e.g. /etc/hosts) modifications
type Hostsfile struct {
	kymaKube kube.KymaKube
}

func NewHostsfile(kymaKube kube.KymaKube) *Hostsfile {
	return &Hostsfile{
		kymaKube: kymaKube,
	}
}

func (h *Hostsfile) UpdateHostFiles() error {
	localDomains, err := h.getLocalDomains()
	if err != nil {
		return err
	}

	//update hosts file
	hostsFile, err := txeh.NewHostsDefault()
	if err != nil {
		return err
	}

	hostsFile.AddHosts(LocalIP, localDomains)
	hostsFile.Save()

	return nil
}

func (h *Hostsfile) getLocalDomains() ([]string, error) {
	domainNames := make([]string, 20)

	vsList, err := h.kymaKube.Istio().NetworkingV1alpha3().VirtualServices("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, v := range vsList.Items {
		for _, host := range v.Spec.Hosts {
			domainNames = append(domainNames, host)
		}
	}

	return domainNames, nil
}
