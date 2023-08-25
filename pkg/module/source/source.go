package source

import (
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type Source interface {
	FetchSource(ctx cpi.Context, path, repo, version string) (*ocm.Source, error)
}
