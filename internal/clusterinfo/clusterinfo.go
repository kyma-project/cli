package clusterinfo

import (
	"context"

	"k8s.io/client-go/kubernetes"
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
