package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// restConfig loads the rest configuration needed by k8s clients to interact with clusters based on the kubeconfig.
// Loading rules are based on standard defined kubernetes config loading.
func restConfig(url, file string) (*rest.Config, error) {
	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommentded file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = file

	cfg, err := clientcmd.BuildConfigFromKubeconfigGetter(url, po.GetStartingConfig)
	if err != nil {
		return nil, err
	}
	cfg.WarningHandler = rest.NoWarnings{}
	return cfg, nil
}

// kubeConfig loads a structured representation of the Kubeconfig.
// Loading rules are based on standard defined kubernetes config loading.
func kubeConfig(file string) (*api.Config, error) {
	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommentded file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = file

	return po.GetStartingConfig()
}

// KubeconfigPath provides the path used to load the kubeconfig based on the standard defined precedence
// if file is a valid path it will be returned, otherwise KUBECONFIG env var or the default location will be used
func KubeconfigPath(file string) string {
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = file

	return po.GetLoadingPrecedence()[0]
}

// AppendConfig adds the provided kubeconfig in the []byte to the Kubeconfig in the target path without altering other existing conifgs.
// If the target path is empty, standard kubeconfig loading rules apply.
func AppendConfig(cfg []byte, target string) error {
	s, err := clientcmd.Load(cfg)
	if err != nil {
		return err
	}

	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommentded file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = target

	t, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	// append contexts
	for k, v := range s.Contexts {
		t.Contexts[k] = v
	}

	// append clusters
	for k, v := range s.Clusters {
		t.Clusters[k] = v
	}

	// append authinfos
	for k, v := range s.AuthInfos {
		t.AuthInfos[k] = v
	}

	t.CurrentContext = s.CurrentContext

	// write config back
	return clientcmd.ModifyConfig(po, *t, false)
}

// RemoveConfig remoes the provided kubeconfig in the []byte from the Kubeconfig in the target path without altering other existing conifgs.
// If the target path is empty, standard kubeconfig loading rules apply.
func RemoveConfig(cfg []byte, target string) error {
	s, err := clientcmd.Load(cfg)
	if err != nil {
		return err
	}

	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommentded file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = target

	t, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	// remove contexts
	for k := range s.Contexts {
		delete(t.Contexts, k)
	}

	// remove clusters
	for k := range s.Clusters {
		delete(t.Clusters, k)
	}

	// remove authinfos
	for k := range s.AuthInfos {
		delete(t.AuthInfos, k)
	}

	t.CurrentContext = ""

	// write config back
	return clientcmd.ModifyConfig(po, *t, false)
}
