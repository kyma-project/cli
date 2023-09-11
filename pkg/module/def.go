package module

import (
	"errors"
	"fmt"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

	"github.com/kyma-project/cli/pkg/module/oci"
)

// Definition contains all infrmation and configuration that defines a module (e.g. component descriptor config, template config, layers, CRs...)
type Definition struct {
	SchemaVersion      string                     // schema version for the ocm descriptor
	Source             string                     // path to the sources to create the module
	Name               string                     // Name of the module (mandatory)
	NameMappingMode    NameMapping                // Component Name mapping as defined in OCM spec.
	Version            string                     // Version of the module (mandatory)
	RegistryURL        string                     // Registry URL to push the image to (optional)
	DefaultCRPath      string                     // path to the file containing the CR to include in the module template  (optional)
	SingleManifestPath string                     // path to the file containing combined manifest for the module (optional)
	Override           bool                       // If true, existing module is overwritten if the configuration differs.
	CustomStateChecks  []v1beta2.CustomStateCheck // specifies optional behavior for determining module state.

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
	return nil
}

func ParseNameMapping(val string) (NameMapping, error) {
	if val == string(URLPathNameMapping) {
		return URLPathNameMapping, nil
	} else if val == string(DigestNameMapping) {
		return DigestNameMapping, nil
	}
	return "", fmt.Errorf(
		"invalid mapping mode: %s, only %s or %s are allowed", val, URLPathNameMapping, DigestNameMapping,
	)
}
