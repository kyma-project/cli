package module

import (
	"fmt"

	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

func InitComponentDescriptor(def *Definition) (*ocm.ComponentDescriptor, error) {
	if err := def.validate(); err != nil {
		return nil, err
	}
	cd := &ocm.ComponentDescriptor{}
	cd.ComponentSpec.SetName(def.Name)
	cd.ComponentSpec.SetVersion(def.Version)
	cd.Metadata.ConfiguredVersion = def.SchemaVersion
	builtByCLI, err := ocmv1.NewLabel("kyma-project.io/built-by", "cli", ocmv1.WithVersion("v1"))
	if err != nil {
		return nil, err
	}
	cd.Provider = ocmv1.Provider{Name: "kyma-project.io", Labels: ocmv1.Labels{*builtByCLI}}

	ocm.DefaultResources(cd)
	if err := ocm.Validate(cd); err != nil {
		return nil, fmt.Errorf("unable to validate component descriptor: %w", err)
	}

	return cd, nil
}
