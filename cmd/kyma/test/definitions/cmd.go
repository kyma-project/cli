package definitions

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	opts *Options
	cli.Command
}

func NewCmd(o *Options) *cobra.Command {
	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:        "definitions",
		Short:      "Shows test definitions available for a provisioned Kyma cluster.",
		Long:       `Use this command to list test definitions available for a provisioned Kyma cluster.`,
		RunE:       func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases:    []string{"def"},
		Deprecated: "`test definitions` is deprecated!",
	}
	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	testDefs, err := listTestDefinitionNames(cmd.K8s.Octopus())
	if err != nil {
		return err
	}
	if len(testDefs) == 0 {
		fmt.Println("No test definitions found")
		return nil
	}
	for _, t := range testDefs {
		fmt.Printf("%s\r\n", t)
	}

	return nil
}

func listTestDefinitionNames(cli octopus.Interface) ([]string, error) {
	defs, err := cli.ListTestDefinitions(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to list test definitions")
	}

	var result = make([]string, len(defs.Items))
	for i := 0; i < len(defs.Items); i++ {
		result[i] = defs.Items[i].GetName()
	}
	return result, nil
}
