package definitions

import (
	"fmt"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *options
	core.Command
}

func NewCmd(o *options) *cobra.Command {
	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "definitions",
		Short:   "Show test definitions available for a provisioned Kyma cluster",
		Long:    `Show test definitions available for a provisioned Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"def"},
	}

	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "could not initialize the Kubernetes client. Make sure that your kubeconfig is valid.")
	}

	if testDefs, err := listTestDefinitionNames(cmd.K8s.Octopus()); err != nil {
		return err
	} else {
		if len(testDefs) == 0 {
			fmt.Println("no test definitions found")
			return nil
		}
		for _, t := range testDefs {
			fmt.Printf("%s\r\n", t)
		}
	}
	return nil
}

func listTestDefinitionNames(cli octopus.OctopusInterface) ([]string, error) {
	defs, err := cli.ListTestDefinitions()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list test definitions")
	}

	var result = make([]string, len(defs.Items))
	for i := 0; i < len(defs.Items); i++ {
		result[i] = defs.Items[i].GetName()
	}
	return result, nil
}
