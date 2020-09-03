package workspace

import (
	"github.com/kyma-project/cli/internal/resources/types"
	"gopkg.in/yaml.v3"
	"io"
)

var _ file = &Cfg{}

const CfgFilename = "config.yaml"

type Cfg struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`

	Runtime    types.Runtime `yaml:"runtime"`
	Git        bool          `yaml:"git,omitempty"`
	SourcePath string        `yaml:"-"`

	Resources  struct {
		Limits ResourceList `yaml:"limits"`
		Requests ResourceList `yaml:"requests"`
	}`yaml:"resources,omitempty"`

	Triggers []struct {
		EventTypeVersion string `yaml:"eventTypeVersion"`
		Source           string `yaml:"source"`
		Type             string `yaml:"type"`
	} `yaml:"triggers,omitempty"`
}

type ResourceList map[ResourceName]string

type ResourceName string

const (
	ResourceCPU ResourceName = "cpu"
	ResourceMemory ResourceName = "memory"
)

func (cfg Cfg) write(writer io.Writer, _ interface{}) error {
	return yaml.NewEncoder(writer).Encode(&cfg)
}

func (cfg Cfg) fileName() string {
	return CfgFilename
}
