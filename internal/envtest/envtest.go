package envtest

import (
	"fmt"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func NewRunner(binPath string, env *envtest.Environment, restClient *rest.Config) *Runner {
	return &Runner{
		binPath, env, restClient,
	}
}

type Runner struct {
	binPath    string
	env        *envtest.Environment
	restClient *rest.Config
}

func (r *Runner) RestClient() *rest.Config {
	return r.restClient
}

func (r *Runner) Start(crdFilePath string, _ *zap.SugaredLogger) (err error) {

	r.env = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{crdFilePath},
		},
		BinaryAssetsDirectory: r.binPath,
		ErrorIfCRDPathMissing: true,
	}

	r.restClient, err = r.env.Start()
	if err != nil {
		return fmt.Errorf("could not start the `envtest` envionment: %w", err)
	}
	if r.restClient == nil {
		return fmt.Errorf("could not get the RestConfig for the `envtest` envionment: %w", err)
	}

	return nil
}

func (r *Runner) Stop() error {
	if err := r.env.Stop(); err != nil {
		return fmt.Errorf("could not stop CR validation: %w", err)
	}
	return nil
}
