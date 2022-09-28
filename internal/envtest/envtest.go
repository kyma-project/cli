package envtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
)

const (
	DefaultVersion  = "1.24.x!" //1.24.x! means the latest available patch version for 1.24 branch
	versionEnv      = "ENVTEST_K8S_VERSION"
	envtestSetupBin = "setup-envtest"
)

func Setup(step step.Step, verbose bool) (string, error) {
	p, err := files.KymaHome()
	if err != nil {
		return "", err
	}

	//Install setup-envtest
	if _, err := os.Stat(filepath.Join(p, envtestSetupBin)); os.IsNotExist(err) {
		if runtime.GOOS == "windows" {
			if _, err := exec.LookPath("bash"); err != nil {
				return "", errors.New("\nBash is not installed. To install bash on windows please see http://win-bash.sourceforge.net")
			}
		}

		kymaGobinEnv := "GOBIN=" + p
		envtestSetupCmd := exec.Command("go", "install", "sigs.k8s.io/controller-runtime/tools/setup-envtest@latest")
		envtestSetupCmd.Env = os.Environ()
		envtestSetupCmd.Env = append(envtestSetupCmd.Env, kymaGobinEnv)

		//go install is silent when executed successfully
		_, err := envtestSetupCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("error installing setup-envtest: %w", err)
		} else if verbose {
			step.LogInfof("Installed setup-envtest in: %q", p)
		}

	}

	//Install envtest binaries using setup-envtest
	envtestSetupBinPath := filepath.Join(p, envtestSetupBin)

	version := os.Getenv(versionEnv)
	if version == "" {
		version = DefaultVersion
	}

	envtestInstallBinariesCmd := exec.Command(envtestSetupBinPath, "use", version, "--bin-dir", p)
	out, err := envtestInstallBinariesCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error installing envtest binaries: %w", err)
	}

	envtestBinariesPath, err := extractPath(string(out))
	if err != nil {
		return "", fmt.Errorf("error installing envtest binaries: %w", err)
	}

	if verbose {
		version, err := extractVersion(string(out))
		if err == nil {
			step.LogInfof("Installed envtest binaries in version %s, path: %q", version, envtestBinariesPath)
		} else {
			step.LogInfof("Installed envtest binaries in: %q", envtestBinariesPath)
		}
	}

	return envtestBinariesPath, nil
}

//extractPath extracts the envtest binaries path from the "setup-envtest" command output
func extractPath(envtestSetupMsg string) (string, error) {
	return parseEnvtestSetupMsg(envtestSetupMsg, `[pP]ath:(.+)`, "envtest binaries path")

}

//extractVersion extracts the envtest binaries version from the "setup-envtest" command output
func extractVersion(envtestSetupMsg string) (string, error) {
	return parseEnvtestSetupMsg(envtestSetupMsg, `[vV]ersion:(.+)`, "envtest version")
}

func parseEnvtestSetupMsg(envtestSetupMsg, rgxp, objName string) (string, error) {
	r, err := regexp.Compile(rgxp)
	if err != nil {
		return "", err
	}
	matches := r.FindStringSubmatch(envtestSetupMsg)
	if len(matches) != 2 {
		return "", fmt.Errorf("Couldn't find %s in the \"setup-envtest\" command output", objName)
	}

	return strings.TrimSpace(matches[1]), nil
}
