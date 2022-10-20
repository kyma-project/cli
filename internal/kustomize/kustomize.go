package kustomize

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
)

const (
	DefaultVersion     = "4.5.7"
	versionEnv         = "KUSTOMIZE_VERSION"
	kustomizeinstaller = "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
	kustomizeBin       = "kustomize"

	buildURLPattern = "%s?ref=%s" // pattern for URL locations Definition.Location?ref=Definition.Ref
	defaultURLRef   = "main"
	localRef        = "local"
)

type Definition struct {
	Name     string
	Ref      string
	Location string
}

func ParseKustomization(s string) (Definition, error) {
	// split URL from ref
	items := strings.Split(s, "@")
	if len(items) == 0 || len(items) > 2 {
		return Definition{}, fmt.Errorf("the given kustomization %q could not be parsed: at least, it must contain a location (URL or path); optionally, URLs can have a reference in format URL@ref", s)
	}

	res := Definition{}
	u, err := url.Parse(items[0])
	if err != nil {
		return Definition{}, fmt.Errorf("could not parse the given location %q: make sure it is a valid URL or path", items[0])
	}

	// URL case
	if u.Scheme != "" && u.Host != "" {
		pathChunks := strings.Split(u.Path, "/")
		if len(pathChunks) < 3 {
			return Definition{}, fmt.Errorf("The provided URL %q does not belong to a repository. It must follow the format DOMAIN.EXT/OWNER/REPO/[SUBPATH]", items[0])
		}
		res.Name = pathChunks[2]
		if len(items) == 2 {
			res.Ref = items[1]
		} else {
			res.Ref = defaultURLRef
		}
		res.Location = items[0]
	} else { // Path case
		res.Name = items[0]
		res.Ref = localRef
		res.Location = items[0]
	}

	return res, nil
}

func Setup(step step.Step, verbose bool) error {
	// check if binary is there (not interested in the path itself at setup)
	_, err := kustomizeBinPath()

	// if not installed, install
	if errors.Is(err, os.ErrNotExist) {
		if runtime.GOOS == "windows" {
			if _, err := exec.LookPath("bash"); err != nil {
				return errors.New("\nBash is not installed. To install bash on windows please see http://win-bash.sourceforge.net")
			}
		}

		v := os.Getenv(versionEnv)
		if v == "" {
			v = DefaultVersion
		}

		home, err := files.KymaHome()
		if err != nil {
			return err
		}

		downloadCmd := exec.Command("curl", "-s", kustomizeinstaller)
		installCmd := exec.Command("bash", "-s", "--", v, home)

		// pipe the downloaded script to the install command
		out, err := cli.Pipe(downloadCmd, installCmd)
		if err != nil {
			return fmt.Errorf("error installing kustomize %w", err)
		} else if verbose {
			step.LogInfof("Installed Kustomize: %s", out)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting kustomize binary: %w", err)
	}
	return nil
}

// Build generates a manifest given a path using kustomize
// Additional args might be given to the kustomize build command
func Build(def Definition, args ...string) ([]byte, error) {
	p, err := kustomizeBinPath()
	if err != nil {
		return nil, fmt.Errorf("error getting kustomize binary: %w", err)
	}

	path := def.Location
	if def.Ref != localRef {
		path = fmt.Sprintf(buildURLPattern, def.Location, def.Ref)
	}

	// prepend command and path to args
	args = append([]string{"build", path}, args...)

	out, err := exec.Command(p, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not build kustomization: %s: %w", out, err)
	}

	return out, nil
}

// Set image edits the given image to the given value in the kustomization found in path.
func SetImage(path, img, value string) error {
	p, err := kustomizeBinPath()
	if err != nil {
		return fmt.Errorf("error getting kustomize binary: %w", err)
	}

	c := exec.Command(p, "edit", "set", "image", fmt.Sprintf("%s=%s", img, value))
	c.Dir = path
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not build kustomization: %s: %w", out, err)
	}

	return nil
}

// kustomizeBinPath looks for the kustomize binary in the PATH or in the default Kyma home folder.
// If it's not there in any location, os.ErrNotExist is returned.
// Any other error means something went wrong.
func kustomizeBinPath() (string, error) {
	p, err := exec.LookPath(kustomizeBin)
	if err != nil && !errors.Is(err, exec.ErrNotFound) {
		return p, err
	}
	if p != "" {
		return p, nil
	}

	home, err := files.KymaHome()
	if err != nil {
		return "", err
	}

	return exec.LookPath(filepath.Join(home, kustomizeBin))
}
