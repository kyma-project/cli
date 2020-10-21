package installation

import (
	"context"
	"strconv"

	"github.com/kyma-project/cli/internal/kube"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterInfo struct {
	IsLocal       bool
	Provider      string
	Profile       string
	LocalIP       string
	LocalVMDriver string
}

func GetClusterInfoFromConfigMap(kymaKube kube.KymaKube) (ClusterInfo, error) {
	cm, err := kymaKube.Static().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kyma-cluster-info", metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return ClusterInfo{}, nil
		}
		return ClusterInfo{}, err
	}

	isLocal, err := strconv.ParseBool(cm.Data["isLocal"])
	if err != nil {
		isLocal = false
	}

	clusterConfig := ClusterInfo{
		IsLocal:       isLocal,
		Provider:      cm.Data["provider"],
		Profile:       cm.Data["profile"],
		LocalIP:       cm.Data["localIP"],
		LocalVMDriver: cm.Data["localVMDriver"],
	}

	return clusterConfig, nil
}
