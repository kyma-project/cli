package certs

import (
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *cli.Options
	cli.Command
}

// NewCmd creates a new dashboard command
func NewCmd(o *cli.Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "certs",
		Short: "Imports certificates in local storage.",
		Long:  `Use this command to add the certificates to the local certificates storage of machine after the installation.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}
	return cmd
}

// Run runs the command
func (cmd *command) Run() error {
	var err error

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	return cmd.importCertificate()
}

func (cmd *command) importCertificate() error {
	f := step.Factory{
		NonInteractive: true,
	}
	s := f.NewStep("Importing Kyma certificate")
	ca := trust.NewCertifier(cmd.K8s)
	if er := root.IsWithSudo(); er != nil {
		s.LogErrorf("%v", er)
		return nil
	}

	// get cert from cluster
	cert, err := ca.CertificateKyma2()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "kyma-*.crt")
	if err != nil {
		return errors.Wrap(err, "cannot create temporary file for Kyma certificate")
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(cert); err != nil {
		return errors.Wrap(err, "failed to write the Kyma certificate")
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// create a simple step to print certificate import steps without a spinner (spinner overwrites sudo prompt)
	// TODO refactor how certifier logs when the old install command is gone
	if err := ca.StoreCertificateKyma2(tmpFile.Name(), s); err != nil {
		return err
	}
	s.Successf("Kyma root certificate imported")
	return nil
}
