package module

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

//go:generate mockery --name OciRepo
type OciRepoImpl interface {
	ComponentVersionExists(archive *comparch.ComponentArchive, repo cpi.Repository) (bool, error)
	GetComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository) (ocm.ComponentVersionAccess, error)
	PushComponentVersion(archive *comparch.ComponentArchive, repository cpi.Repository, overwrite bool) error
}

type OciRepo struct{}

func (r *OciRepo) ComponentVersionExists(archive *comparch.ComponentArchive, repo cpi.Repository) (bool, error) {
	return repo.ExistsComponentVersion(archive.ComponentVersionAccess.GetName(),
		archive.ComponentVersionAccess.GetVersion())
}

func (r *OciRepo) GetComponentVersion(archive *comparch.ComponentArchive,
	repo cpi.Repository) (ocm.ComponentVersionAccess, error) {
	return repo.LookupComponentVersion(archive.ComponentVersionAccess.GetName(),
		archive.ComponentVersionAccess.GetVersion())
}

func (r *OciRepo) PushComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository, overwrite bool) error {
	transferHandler, err := standard.New(standard.Overwrite(overwrite))
	if err != nil {
		return fmt.Errorf("could not setup archive transfer: %w", err)
	}

	if err = transfer.TransferVersion(
		common.NewLoggingPrinter(archive.GetContext().Logger()), nil, archive.ComponentVersionAccess, repo,
		&customTransferHandler{transferHandler},
	); err != nil {
		return fmt.Errorf("could not finish component transfer: %w", err)
	}
	return nil
}
