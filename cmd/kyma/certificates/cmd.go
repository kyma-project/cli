package certificates

import (
	"io/ioutil"
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

//NewCmd creates a new dashboard command
func NewCmd(o *cli.Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "add-certificates",
		Short: "Adds certificates to local storage.",
		Long:  `Use this command to add the certificates to the local storage of machine after the installation.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}
	return cmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	if err := cmd.importCertificate(); err != nil {
		return err
	}
	return nil
}

func (cmd *command) importCertificate() error {
	f := step.Factory{
		NonInteractive: true,
	}
	s := f.NewStep("Importing Kyma certificate")

	if !root.IsWithSudo() {
		s.LogError("Could not store certificates locally. Make sure you are using sudo.")
		return nil
	}

	ca := trust.NewCertifier(cmd.K8s)

	if !cmd.approveImportCertificate() {
		//no approval given: stop import
		ca.InstructionsKyma2()
		return nil
	}

	// get cert from cluster
	cert, err := ca.CertificateKyma2()
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "kyma-*.crt")
	if err != nil {
		return errors.Wrap(err, "Cannot create temporary file for Kyma certificate.")
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(cert); err != nil {
		return errors.Wrap(err, "Failed to write the Kyma certificate.")
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// create a simple step to print certificate import steps without a spinner (spinner overwrites sudo prompt)
	// TODO refactor how certifier logs when the old install command is gone
	if err := ca.StoreCertificate(tmpFile.Name(), s); err != nil {
		return err
	}
	s.Successf("Kyma root certificate imported")
	return nil
}

func (cmd *command) approveImportCertificate() bool {
	qImportCertsStep := cmd.NewStep("Install Kyma certificate locally")
	defer qImportCertsStep.Success()
	if cmd.avoidUserInteraction() {
		//do not import if user-interaction has to be avoided (suppress sudo pwd request)
		return false
	}
	return qImportCertsStep.PromptYesNo("Do you want to install the Kyma certificate locally?")
}

func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}
