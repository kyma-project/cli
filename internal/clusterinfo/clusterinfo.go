package clusterinfo

import (
	"context"
	"k8s.io/client-go/kubernetes"
)

type Info struct {
	//ClusterType represent a type of the cluster (k3d, gardener, etc.)
	ClusterType ClusterType

	//Domain represents the cluster domain (currently only works on Gardener)
	Domain string

	//ClusterName represents the name of the cluster, which every node will be prefixed with (currently only works on k3d)
	ClusterName string
}

func Get(ctx context.Context, kubeClient kubernetes.Interface) (*Info, error) {
	gardenerDomain, err := getGardenerDomain(ctx, kubeClient)
	if err != nil {
		return nil, err
	}

	return &Info{
		ClusterType: Gardener,
		Domain:      gardenerDomain,
	}, nil
}
