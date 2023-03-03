package module

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/ocm.software/v3alpha1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

// nolint:gochecknoinits
func init() {
	compdesc.RegisterScheme(&v3alpha1.DescriptorVersion{})
	ocm.DefaultContext().RepositoryTypes().Register(ocireg.Type, genericocireg.NewRepositoryType(oci.DefaultContext()))
	ocm.DefaultContext().RepositoryTypes().Register(
		ocireg.LegacyType, genericocireg.NewRepositoryType(oci.DefaultContext()),
	)
	ocm.DefaultContext().RepositoryTypes().Register(
		comparch.Type, cpi.NewRepositoryType(comparch.Type, &comparch.RepositorySpec{}, nil),
	)
}
