package envtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kyma-project/cli/internal/envtest"
	"github.com/kyma-project/cli/internal/files"
)

const (
	envtestDefaultVersion = "1.24.x!" //1.24.x! means the latest available patch version for 1.24 branch
	envtestVersionEnv     = "ENVTEST_K8S_VERSION"
	envtestSetupBin       = "setup-envtest"
)

// based on "kubernetes-sigs/controller-runtime/tools/setup-envtest/versions/parse.go", but more strict
var envtestVersionRegexp = regexp.MustCompile(`^(0|[1-9]\d{0,2})\.(0|[1-9]\d{0,2})\.(0|[1-9]\d{0,3})$`)

func EnvTest() (*envtest.Runner, error) {
	p, err := files.KymaHome()
	if err != nil {
		return nil, err
	}

	//Install setup-envtest
	if _, err := os.Stat(filepath.Join(p, envtestSetupBin)); os.IsNotExist(err) {
		kymaGobinEnv := "GOBIN=" + p
		envtestSetupCmd := exec.Command("go", "install", "sigs.k8s.io/controller-runtime/tools/setup-envtest@latest")
		envtestSetupCmd.Env = os.Environ()
		envtestSetupCmd.Env = append(envtestSetupCmd.Env, kymaGobinEnv)

		//go install is silent when executed successfully
		out, err := envtestSetupCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error installing setup-envtest: %w. Details: %s", err, string(out))
		}

	}

	//Install envtest binaries using setup-envtest
	envtestSetupBinPath := filepath.Join(p, envtestSetupBin)

	version, err := resolveEnvtestVersion()
	if err != nil {
		return nil, err
	}

	envtestInstallBinariesCmd := exec.Command(envtestSetupBinPath, "use", version, "--bin-dir", p)
	out, err := envtestInstallBinariesCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error installing envtest binaries: %w. Details: %s", err, string(out))
	}

	envtestBinariesPath, err := extractPath(string(out))
	if err != nil {
		return nil, fmt.Errorf("error installing envtest binaries: %w", err)
	}

	return envtest.NewRunner(envtestBinariesPath, nil, nil), nil
}

// resolveEnvtestVersion validates the envtest version provided via the environment variable. It returns the default version if the variable is not found.
func resolveEnvtestVersion() (string, error) {
	v, defined := os.LookupEnv(envtestVersionEnv)
	if !defined {
		return envtestDefaultVersion, nil
	}

	trimmed := strings.TrimSpace(v)
	if !envtestVersionRegexp.MatchString(trimmed) {
		return "", errors.New("Invalid value of \"ENVTEST_K8S_VERSION\" variable, only proper semversions are allowed, e.g: 1.24.2")
	}

	return trimmed, nil
}

// extractPath extracts the envtest binaries path from the "setup-envtest" command output
func extractPath(envtestSetupMsg string) (string, error) {
	return parseEnvtestSetupMsg(envtestSetupMsg, `[pP]ath:(.+)`, "envtest binaries path")

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
