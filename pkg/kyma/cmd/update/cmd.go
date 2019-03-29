package update

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	core.Command
}

const (
	sleep = 10 * time.Second
)

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "update",
		Short: "Updates an already installed Kyma",
		Long: `Updates an already installed Kyma.

Assure that your KUBECONFIG is pointing to the target cluster already.
The command will:
- Update the version of the kyma-installer tiller
- Triggers the installation
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().StringVarP(&o.ReleaseVersion, "release", "r", "0.8.2", "kyma release to use")
	cobraCmd.Flags().BoolVarP(&o.NoWait, "noWait", "n", false, "Do not wait for completion of kyma-installer")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 0, "Timeout after which CLI should give up watching installation")

	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	s := cmd.NewStep("Checking requirements")
	err := cmd.checkRequirements()
	if err != nil {
		s.Failure()
		return err
	}

	success, err := cmd.checkVersion()
	if err != nil {
		s.Failure()
		return err
	}
	if success != "" {
		s.Successf(success)
		return nil
	}
	s.Successf("Requirements are fine")

	s = cmd.NewStep("Updating kyma-installer")
	err = cmd.updateInstaller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer updated")

	s = cmd.NewStep("Requesting kyma-installer to update kyma")
	err = cmd.activateInstaller()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("kyma-installer is updating kyma")

	if !cmd.opts.NoWait {
		err = cmd.waitForInstaller()
		if err != nil {
			return err
		}
	}

	err = cmd.printSummary()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) checkRequirements() error {
	versionWarning, err := cmd.Kubectl().CheckVersion()
	if err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	if versionWarning != "" {
		cmd.CurrentStep.LogError(versionWarning)
	}

	return nil
}

func (cmd *command) checkVersion() (string, error) {
	version, err := internal.GetKymaVersion(cmd.opts.Verbose)
	if err != nil {
		return "", err
	}
	if version == cmd.opts.ReleaseVersion {
		cmd.CurrentStep.Success()
		return "Version is up to date already, nothing to do", nil
	}

	var answer string
	if !cmd.opts.NonInteractive {
		answer, err = cmd.CurrentStep.Prompt("Installed Kyma version is " + version + ". Do you want to update your installation to Kyma version " + cmd.opts.ReleaseVersion + " [y/N]: ")
		if err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(answer) != "y" {
		return "Cancelled update process", nil
	}
	return "", nil
}

func (cmd *command) updateInstaller() error {
	_, err := cmd.Kubectl().RunCmd("-n", "kyma-installer", "patch", "deployment", "kyma-installer", "--type=json", fmt.Sprintf("-p=[{'op': 'replace', 'path': '/spec/template/spec/containers/0/image', 'value':'eu.gcr.io/kyma-project/kyma-installer:%s'}]", cmd.opts.ReleaseVersion), "--v=8")
	if err != nil {
		return err
	}

	err = cmd.Kubectl().WaitForPodReady("kyma-installer", "name", "kyma-installer")
	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) activateInstaller() error {
	status, err := cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "InProgress" {
		return nil
	}

	_, err = cmd.Kubectl().RunCmd("label", "installation/kyma-installation", "action=install")
	if err != nil {
		return err
	}
	return nil
}

func (cmd *command) printSummary() error {
	version, err := internal.GetKymaVersion(cmd.opts.Verbose)
	if err != nil {
		return err
	}

	clusterInfo, err := cmd.Kubectl().RunCmd("cluster-info")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(clusterInfo)
	fmt.Println()
	fmt.Printf("Kyma is updated to version %s\n", version)
	fmt.Println()
	fmt.Println("Happy Kyma-ing! :)")
	fmt.Println()

	return nil
}

func (cmd *command) waitForInstaller() error {
	currentDesc := ""
	_ = cmd.NewStep("Waiting for installation to start")

	status, err := cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return err
	}
	if status == "Installed" {
		return nil
	}

	var timeout <-chan time.Time
	var errorOccured bool
	if cmd.opts.Timeout > 0 {
		timeout = time.After(cmd.opts.Timeout)
	}

	for {
		select {
		case <-timeout:
			cmd.CurrentStep.Failure()
			_ = cmd.printInstallationErrorLog()
			return errors.New("Timeout while awaiting installation to complete")
		default:
			status, desc, err := cmd.getInstallationStatus()
			if err != nil {
				return err
			}

			switch status {
			case "Installed":
				cmd.CurrentStep.Success()
				return nil

			case "Error":
				if !errorOccured {
					errorOccured = true
					cmd.CurrentStep.Failuref("Error installing Kyma: %s", desc)
					cmd.CurrentStep.LogInfof("To fetch the logs from the installer execute: 'kubectl logs -n kyma-installer -l name=kyma-installer'")
				}

			case "InProgress":
				errorOccured = false
				// only do something if the description has changed
				if desc != currentDesc {
					cmd.CurrentStep.Success()
					cmd.CurrentStep = cmd.opts.NewStep(fmt.Sprintf(desc))
					currentDesc = desc
				}

			default:
				cmd.CurrentStep.Failure()
				fmt.Printf("Unexpected status: %s\n", status)
				os.Exit(1)
			}
			time.Sleep(sleep)
		}
	}
}

func (cmd *command) getInstallationStatus() (status string, desc string, err error) {
	status, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return
	}
	desc, err = cmd.Kubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'")
	return
}

func (cmd *command) printInstallationErrorLog() error {
	logs, err := cmd.Kubectl().RunCmd("get", "installation", "kyma-installation", "-o", "go-template", `--template={{- range .status.errorLog -}}
{{.component}}:
{{.log}} [{{.occurrences}}]

{{- end}}
`)
	if err != nil {
		return err
	}
	fmt.Println(logs)
	return nil
}
