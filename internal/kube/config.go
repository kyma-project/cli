package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// restConfig loads the rest configuration needed by k8s clients to interact with clusters based on the kubeconfig.
// Loading rules are based on standard defined kubernetes config loading.
func restConfig(apiConfig *api.Config) (*rest.Config, error) {
	cfg, err := clientcmd.NewDefaultClientConfig(*apiConfig, nil).ClientConfig()
	if err != nil {
		return nil, err
	}

	cfg.WarningHandler = rest.NoWarnings{}
	return cfg, nil
}

// apiConfig loads a structured representation of the Kubeconfig.
// Loading rules are based on standard defined kubernetes config loading.
func apiConfig(kubeconfig, context string) (*api.Config, error) {
	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommended file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = kubeconfig

	api, err := po.GetStartingConfig()

	if context != "" {
		api.CurrentContext = context
	}

	return api, err
}

// setKubernetesDefaults sets default values on the provided client config for accessing the
// Kubernetes API or returns an error if any of the defaults are impossible or invalid.
func setKubernetesDefaults(config *rest.Config) error {
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}

	if config.APIPath == "" {
		config.APIPath = "/api"
	}

	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	}

	return rest.SetKubernetesDefaults(config)
}

// SaveConfig saves the kubeconfig to a file or prints it to the console.
func SaveConfig(kubeconfig *api.Config, output string) error {
	if output != "" {
		err := clientcmd.WriteToFile(*kubeconfig, output)
		if err != nil {
			return err
		}
		println("Kubeconfig saved to: " + output)
		return nil
	}
	message, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return err
	}
	fmt.Println(string(message))
	return nil
}
