package kustomize

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
)

const (
	DefaultVersion     = "4.5.7"
	versionEnv         = "KUSTOMIZE_VERSION"
	kustomizeinstaller = "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
	kustomizeBin       = "kustomize"
)

func Setup(step step.Step, verbose bool) error {
	p, err := files.KymaHome()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(p, kustomizeBin)); os.IsNotExist(err) {
		if runtime.GOOS == "windows" {
			if _, err := exec.LookPath("bash"); err != nil {
				return errors.New("\nBash is not installed. To install bash on windows please see http://win-bash.sourceforge.net")
			}
		}

		v := os.Getenv(versionEnv)
		if v == "" {
			v = DefaultVersion
		}

		downloadCmd := exec.Command("curl", "-s", kustomizeinstaller)
		installCmd := exec.Command("bash", "-s", "--", v, p)

		// pipe the downloaded script to the install command
		out, err := cli.Pipe(downloadCmd, installCmd)
		if err != nil {
			return fmt.Errorf("error installing kustomize %w", err)
		} else if verbose {
			step.LogInfof("Installed Kustomize: %s", out)
		}
	}
	return nil
}

// Build generates a manifest given a path using kustomize
func Build(path string) ([]byte, error) {
	p, err := kustomizeBinPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(p, "build", path)

	return cmd.CombinedOutput()
}

// Set edits the kustomize file in the given path by setting the given key to the given value in the given resource.
// Resource can be one of: annotation, buildmetadata, image, label, nameprefix, namespace, namesuffix, replicas.
func Set(path, resource, key, value string) error {
	p, err := kustomizeBinPath()
	if err != nil {
		return err
	}

	cmd := exec.Command(p, "edit", "set", key)
	cmd.Path = path

	return nil
}

func kustomizeBinPath() (string, error) {
	p, err := files.KymaHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(p, kustomizeBin), nil
}
