package module

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	// nolint:revive
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

// nolint:gochecknoinits
func init() {
	ocm.DefaultContext().RepositoryTypes().Register(ocireg.Type, genericocireg.NewRepositoryType(oci.DefaultContext()))
	ocm.DefaultContext().RepositoryTypes().Register(
		ocireg.LegacyType, genericocireg.NewRepositoryType(oci.DefaultContext()),
	)
	ocm.DefaultContext().RepositoryTypes().Register(
		comparch.Type, cpi.NewRepositoryType(comparch.Type, &comparch.RepositorySpec{}, nil),
	)
}
