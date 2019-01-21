package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigFactory struct {
	KubeconfigPath string
	kubeconfig *rest.Config
}

func (f *ConfigFactory) GetKubeconfig() (*rest.Config, error) {
	if f.kubeconfig != nil {
		return f.kubeconfig, nil
	}

	config, err := clientcmd.BuildConfigFromFlags("", f.KubeconfigPath)
	if err != nil {
		return &rest.Config{}, err
	}

	f.kubeconfig = config
	return config, nil
}
