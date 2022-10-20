package kubebuilder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/cli/internal/kustomize"
	"github.com/kyma-project/cli/pkg/step"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/yaml"
)

const (
	V3      = "go.kubebuilder.io/v3"
	V4alpha = "go.kubebuilder.io/v4alpha"

	projectFile          = "operator/PROJECT"
	configFile           = "config.yaml"
	defaultKustomization = "operator/config/default"
	samplesPath          = "operator/config/samples/"

	crdFileIdentifier = "customresourcedefinition"
	chartsFolder      = "charts/%s-operator"
	templatesFolder   = "templates"
	crdsFolder        = "crds"
)

type Project struct {
	Layout []string `json:"layout,omitempty"`
	Name   string   `json:"projectName,omitempty"`
	Repo   string   `json:"repo,omitempty"`
	path   string
}

// ParseProject parses the given kubebuilder project and returns a type containing its metadata and with methods to execute kustomize actions on the project
func ParseProject(path string) (*Project, error) {
	yml, err := os.ReadFile(filepath.Join(path, projectFile))
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := yaml.Unmarshal(yml, p); err != nil {
		return nil, fmt.Errorf("could not parse project file: %w", err)
	}

	p.path = path
	return p, nil
}

// Build builds the kubebuilder project default kustomization in the given path. Sets the image the tag: <registry>/<name>-operator:<version>; and returns the folder containing the resulting chart.
func (p *Project) Build(name, version, registry string) (string, error) {
	// check layout
	if !(slices.Contains(p.Layout, V3) || slices.Contains(p.Layout, V4alpha)) {
		return "", fmt.Errorf("project layout %v is not supported", p.Layout)
	}
	// edit kustomization image and setup build
	img := ""
	if registry == "" {
		img = fmt.Sprintf("%s-operator:%s", name, version)
	} else {
		img = fmt.Sprintf("%s/%s-operator:%s", registry, name, version)
	}

	buildPath := filepath.Join(p.path, defaultKustomization)
	if err := kustomize.SetImage(buildPath, "controller", img); err != nil {
		return "", fmt.Errorf("could not edit kustomization image: %w", err)
	}
	k, err := kustomize.ParseKustomization(buildPath)
	if err != nil {
		return "", err
	}

	// create output folders
	chartsPath := filepath.Join(p.path, fmt.Sprintf(chartsFolder, name))
	outPath := filepath.Join(chartsPath, templatesFolder)
	crdsPath := filepath.Join(chartsPath, crdsFolder)

	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create chart templates output dir: %w", err)
	}
	if err := os.MkdirAll(crdsPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create chart CRDs output dir: %w", err)
	}

	// do build
	if _, err := kustomize.Build(k, "-o", outPath); err != nil {
		return "", err
	}

	// move CRDs to their folder
	mvFn := func(path string, d fs.DirEntry, err error) error {
		fileName := filepath.Base(path)
		if strings.Contains(fileName, crdFileIdentifier) {
			if err := os.Rename(path, filepath.Join(crdsPath, fileName)); err != nil {
				return fmt.Errorf("could not move CRD file from %q to %q: %w", path, crdsPath, err)
			}
		}
		return nil
	}

	if err := filepath.WalkDir(outPath, mvFn); err != nil {
		return "", err
	}

	return chartsPath, nil
}

func (p *Project) Config() (string, error) {
	configPath := filepath.Join(p.path, configFile)
	info, err := os.Stat(configPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("expected file but found directory at %q", configPath)
	}
	return configPath, nil
}

// DefaultCR checks the samples of the project to obtain the default CR for the operator and returns its contents.
// Should there be several sample files, the user will be asked to specify which one to use.
func (p *Project) DefaultCR(s step.Step) ([]byte, error) {
	// check layout
	if !(slices.Contains(p.Layout, V3) || slices.Contains(p.Layout, V4alpha)) {
		return nil, fmt.Errorf("project layout %v is not supported", p.Layout)
	}

	samplesDir := filepath.Join(p.path, samplesPath)
	d, err := os.ReadDir(samplesDir)
	if err != nil {
		return nil, fmt.Errorf("could not read samples dir %q: %w", samplesDir, err)
	}

	if len(d) == 0 {
		return nil, fmt.Errorf("no default CR available: samples directory %q is empty", samplesDir)
	}
	defaultCR := ""
	if len(d) > 1 {
		// ask for specific file
		names := []string{}
		for _, f := range d {
			names = append(names, f.Name())
		}

		answer, err := s.Prompt(fmt.Sprintf("Please specify the file to use as default CR in %s: %v\n", samplesDir, names))
		defaultCR = filepath.Join(samplesDir, answer)
		if err != nil {
			return nil, fmt.Errorf("could not obtain default CR from user prompt: %w", err)
		}
	} else {
		// use only file in folder
		defaultCR = filepath.Join(samplesDir, d[0].Name())
	}

	return os.ReadFile(defaultCR)
}
