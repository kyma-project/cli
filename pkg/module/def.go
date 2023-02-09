package module

import (
	"errors"
	"fmt"

	"github.com/kyma-project/cli/pkg/module/oci"
)

// Definition contains all infrmation and configuration that defines a module (e.g. component descriptor config, template config, layers, CRs...)
type Definition struct {
	Source          string      // path to the sources to create the module
	ArchivePath     string      // Location of the component descriptor and the archive to create the module image. If it does not exist, it is created.
	Name            string      // Name of the module (mandatory)
	NameMappingMode NameMapping // Component Name mapping as defined in OCM spec.
	Version         string      // Version of the module (mandatory)
	RegistryURL     string      // Registry URL to push the image to (optional)
	DefaultCRPath   string      // path to the file containing the CR to include in the module template  (optional)
	Overwrite       bool        // If true, existing module is overwritten if the configuration differs.

	// these fields will be filled out when inspecting the module contents
	Layers    []Layer
	Repo      string
	DefaultCR []byte
}

// validate checks that the configuration has all required data for a module to be valid.
func (cfg *Definition) validate() error {
	if cfg.Name == "" {
		return errors.New("The module name cannot be empty")
	}

	ref, err := oci.ParseRef(cfg.Name)
	if err != nil {
		return err
	}

	if err := ValidateName(ref.ShortName()); err != nil {
		return err
	}

	if cfg.Version == "" {
		return errors.New("The module version cannot be empty")
	}
	if cfg.Source == "" {
		return errors.New("The module source path cannot be empty")
	}
	if cfg.ArchivePath == "" {
		return errors.New("The module archive path cannot be empty")
	}
	return nil
}

func ParseNameMapping(val string) (NameMapping, error) {
	if val == string(URLPathNameMapping) {
		return URLPathNameMapping, nil
	} else if val == string(DigestNameMapping) {
		return DigestNameMapping, nil
	}
	return "", fmt.Errorf("invalid mapping mode: %s, only %s or %s are allowed", val, URLPathNameMapping, DigestNameMapping)
}
