package provision

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	hf "github.com/kyma-project/hydroform/provision"
	"github.com/kyma-project/hydroform/provision/types"
)

type Command interface {
	// getters
	IsVerbose() bool
	KubeconfigPath() string
	Attempts() uint
	ProviderName() string

	ValidateFlags() error
	NewStep(msg string) step.Step
	NewCluster() *types.Cluster
	NewProvider() (*types.Provider, error)

	Run() error
}

func RunTemplate(c Command) error {
	s := c.NewStep("Validating flags")
	if err := c.ValidateFlags(); err != nil {
		s.Failure()
		return err
	}
	s.Success()

	cluster := c.NewCluster()
	provider, err := c.NewProvider()
	if err != nil {
		return err
	}
	if !c.IsVerbose() {
		// discard all the noise from terraform logs if not verbose
		log.SetOutput(ioutil.Discard)
	}
	s = c.NewStep(fmt.Sprintf("Provisioning %s cluster", c.ProviderName()))
	home, err := files.KymaHome()
	if err != nil {
		s.Failure()
		return err
	}

	err = retry.Do(
		func() error {
			cluster, err = hf.Provision(cluster, provider, types.WithDataDir(home), types.Persistent(), types.Verbose(c.IsVerbose()))
			return err
		},
		retry.Attempts(c.Attempts()), retry.LastErrorOnly(!c.IsVerbose()))

	if err != nil {
		s.Failure()
		return err
	}
	s.Success()

	s = c.NewStep("Importing kubeconfig")
	kubeconfig, err := hf.Credentials(cluster, provider, types.WithDataDir(home), types.Persistent(), types.Verbose(c.IsVerbose()))
	if err != nil {
		s.Failure()
		return err
	}

	if err := kube.AppendConfig(kubeconfig, c.KubeconfigPath()); err != nil {
		s.Failure()
		return err
	}
	s.Success()

	fmt.Printf("\n%s cluster installed\nKubectl correctly configured: pointing to %s\n\nHappy %s-ing! :)\n", c.ProviderName(), cluster.Name, c.ProviderName())
	return nil
}
