package clusterinfo

import (
	"context"
	"fmt"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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

// IsManagedKyma returns true if the k8s go-client is configured to access a managed Kyma runtime
func IsManagedKyma(ctx context.Context, restConfig *rest.Config, kubeClient kubernetes.Interface) (bool, error) {
	if strings.HasSuffix(restConfig.Host, legacyEnvAPIServerSuffix) {
		return lookupConfigMapMarker(ctx, kubeClient)
	}
	if strings.HasSuffix(restConfig.Host, skrEnvAPIServerSuffix) {
		return lookupConfigMapMarker(ctx, kubeClient)
	}

	return false, nil
}

// lookupConfigMapMarker tries to find a "kyma-system/skr-configmap" marker ConfigMap with specific labels and payload.
func lookupConfigMapMarker(ctx context.Context, kubeClient kubernetes.Interface) (bool, error) {
	cm, err := kubeClient.CoreV1().ConfigMaps("kyma-system").Get(ctx, "skr-configmap", metav1.GetOptions{})

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get ConfigMap \"skr-configmap\"  in the \"kyma-system\" namespace: %w", err)
	}

	if cm == nil || cm.ObjectMeta.Labels == nil || cm.Data == nil {
		return false, nil
	}

	res := cm.ObjectMeta.Labels["reconciler.kyma-project.io/managed-by"] == "reconciler" &&
		cm.Data["is-managed-kyma-runtime"] == "true"

	return res, nil
}
