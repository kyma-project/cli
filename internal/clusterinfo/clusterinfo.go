package clusterinfo

import "k8s.io/client-go/kubernetes"

type Info struct {
	Provider    ClusterProvider
	Domain      string
	ClusterName string
}

func Get(kubeClient kubernetes.Interface) Info {

}
