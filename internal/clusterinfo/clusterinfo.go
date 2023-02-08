package clusterinfo

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	legacyEnvAPIServerSuffix = ".k8s-hana.ondemand.com"
	skrEnvAPIServerSuffix    = ".kyma.ondemand.com" //works for DEV,STAGE and PROD
)

// Info is a discriminated union (can be either Gardener or K3d or Unrecognized)
type Info interface {
	//unexported method to make sure that Info members are only implemented by the clusterinfo package
	sealed()
}

type Gardener struct {
	Domain string
}

func (Gardener) sealed() {}

type K3d struct {
	ClusterName string
}

func (K3d) sealed() {}

type GKE struct{}

func (GKE) sealed() {}

type Unrecognized struct {
}

func (Unrecognized) sealed() {}

func Discover(ctx context.Context, kubeClient kubernetes.Interface) (Info, error) {
	gardenerDomain, err := getGardenerDomain(ctx, kubeClient)
	if err != nil {
		return nil, err
	}

	if gardenerDomain != "" {
		return Gardener{Domain: gardenerDomain}, nil
	}

	isGke, err := isGkeCluster(ctx, kubeClient)
	if err != nil {
		return nil, err
	}
	if isGke {
		return GKE{}, nil
	}

	isK3d, err := isK3dCluster(ctx, kubeClient)
	if err != nil {
		return nil, err
	}

	if !isK3d {
		return Unrecognized{}, nil
	}

	k3dClusterName, err := k3dClusterName(ctx, kubeClient)
	if err != nil {
		return nil, err
	}

	return K3d{ClusterName: k3dClusterName}, nil
}

// IsManagedKyma returns true if the k8s go-client is configured to access a managed kyma runtime
func IsManagedKyma(ctx context.Context, restConfig *rest.Config, kubeClient kubernetes.Interface) (bool, error) {
	//1) Verfiy the domain
	if strings.HasSuffix(restConfig.Host, legacyEnvAPIServerSuffix) {
		//Legacy, may be removed in the future
		return lookupConfigMapMarker(ctx, kubeClient)
	}
	if strings.HasSuffix(restConfig.Host, skrEnvAPIServerSuffix) {
		return lookupConfigMapMarker(ctx, kubeClient)
	}

	return false, nil
}

func lookupConfigMapMarker(ctx context.Context, kubeClient kubernetes.Interface) (bool, error) {
	opts := metav1.ListOptions{LabelSelector: "reconciler.kyma-project.io/managed-by=reconciler"}
	cmList, err := kubeClient.CoreV1().ConfigMaps("kyma-system").List(ctx, opts)
	if err != nil {
		return false, fmt.Errorf("Error listing ConfigMaps in the \"kyma-system\" namespace: %w", err)
	}

	for _, cm := range cmList.Items {
		if cm.Data["is-managed-kyma-runtime"] == "true" {
			return true, nil
		}
	}

	return false, nil
}
