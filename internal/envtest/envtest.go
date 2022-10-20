package envtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	DefaultVersion  = "1.24.x!" //1.24.x! means the latest available patch version for 1.24 branch
	versionEnv      = "ENVTEST_K8S_VERSION"
	envtestSetupBin = "setup-envtest"
)

// based on "kubernetes-sigs/controller-runtime/tools/setup-envtest/versions/parse.go", but more strict
var envtestVersionRegexp = regexp.MustCompile(`^(0|[1-9]\d{0,2})\.(0|[1-9]\d{0,2})\.(0|[1-9]\d{0,3})$`)

type Runner struct {
	binPath    string
	env        *envtest.Environment
	restClient *rest.Config
}

func Setup(step step.Step, verbose bool) (*Runner, error) {
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
		} else if verbose {
			step.LogInfof("Installed setup-envtest in: %q", p)
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

	if verbose {
		version, err := extractVersion(string(out))
		if err == nil {
			step.LogInfof("Installed envtest binaries in version %s, path: %q", version, envtestBinariesPath)
		} else {
			step.LogInfof("Installed envtest binaries in: %q", envtestBinariesPath)
		}
	}

	return &Runner{
		binPath: envtestBinariesPath,
	}, nil
}

func (r *Runner) RestClient() *rest.Config {
	return r.restClient
}

func (r *Runner) Start(crdFilePath string, log *zap.SugaredLogger) (err error) {

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

// extractPath extracts the envtest binaries path from the "setup-envtest" command output
func extractPath(envtestSetupMsg string) (string, error) {
	return parseEnvtestSetupMsg(envtestSetupMsg, `[pP]ath:(.+)`, "envtest binaries path")

}

// extractVersion extracts the envtest binaries version from the "setup-envtest" command output
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

// resolveEnvtestVersion validates the envtest version provided via the environment variable. It returns the default version if the variable is not found.
func resolveEnvtestVersion() (string, error) {
	v, defined := os.LookupEnv(versionEnv)
	if !defined {
		return DefaultVersion, nil
	}

	trimmed := strings.TrimSpace(v)
	if !envtestVersionRegexp.MatchString(trimmed) {
		return "", errors.New("Invalid value of \"ENVTEST_K8S_VERSION\" variable, only proper semversions are allowed, e.g: 1.24.2")
	}

	return trimmed, nil
}
