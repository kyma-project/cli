package module

import (
	"fmt"
	"reflect"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/runtime"
)

//go:generate mockery --name OciRepoAccess --replace-type github.com/open-component-model/ocm/pkg/contexts/ocm/internal=ocm:github.com/open-component-model/ocm/pkg/contexts/ocm
type OciRepoAccess interface {
	ComponentVersionExists(archive *comparch.ComponentArchive, repo cpi.Repository) (bool, error)
	GetComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository) (ocm.ComponentVersionAccess, error)
	PushComponentVersion(archive *comparch.ComponentArchive, repository cpi.Repository, overwrite bool) error
	DescriptorResourcesAreEquivalent(archive *comparch.ComponentArchive, remoteVersion ocm.ComponentVersionAccess) bool
}

type OciRepo struct{}

func (r *OciRepo) ComponentVersionExists(archive *comparch.ComponentArchive, repo cpi.Repository) (bool, error) {
	return repo.ExistsComponentVersion(archive.GetName(),
		archive.GetVersion())
}

func (r *OciRepo) GetComponentVersion(archive *comparch.ComponentArchive,
	repo cpi.Repository) (ocm.ComponentVersionAccess, error) {
	return repo.LookupComponentVersion(archive.GetName(),
		archive.GetVersion())
}

func (r *OciRepo) PushComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository, overwrite bool) error {
	transferHandler, err := standard.New(standard.Overwrite(overwrite))
	if err != nil {
		return fmt.Errorf("could not setup archive transfer: %w", err)
	}

	if err = transfer.TransferVersion(
		common.NewLoggingPrinter(archive.GetContext().Logger()), nil, archive, repo,
		&customTransferHandler{transferHandler},
	); err != nil {
		return fmt.Errorf("could not finish component transfer: %w", err)
	}
	return nil
}

func (r *OciRepo) DescriptorResourcesAreEquivalent(archive *comparch.ComponentArchive,
	remoteVersion ocm.ComponentVersionAccess) bool {
	localResources := archive.GetDescriptor().Resources
	remoteResources := remoteVersion.GetDescriptor().Resources
	if len(localResources) != len(remoteResources) {
		return false
	}

	localResourcesMap := map[string]compdesc.Resource{}
	for _, res := range localResources {
		localResourcesMap[res.Name] = res
	}

	for _, res := range remoteResources {
		localResource := localResourcesMap[res.Name]
		if res.Name == RawManifestLayerName {
			remoteAccess, ok := res.Access.(*runtime.UnstructuredVersionedTypedObject)
			if !ok {
				return false
			}

			_, ok = localResourcesMap[res.Name]
			if !ok {
				return false
			}
			localAccessObject, ok := localResource.Access.(*localblob.AccessSpec)
			if !ok {
				return false
			}

			remoteAccessLocalReference, ok := remoteAccess.Object[accessLocalReferenceFieldName].(string)
			if !ok {
				return false
			}
			// Trimming 7 characters because locally the sha256 is followed by '.' but remote it is followed by ':'
			if remoteAccessLocalReference[7:] != localAccessObject.LocalReference[7:] {
				return false
			}
		} else if !isEquivalent(&res, &localResource) {
			return false
		}
	}

	return true
}

func isEquivalent(r *compdesc.Resource, e compdesc.ElementMetaAccessor) bool {
	// Paranoid sanity checks
	if r == nil && e == nil {
		return true
	}
	if r == nil && e != nil || r != nil && e == nil {
		return false
	}

	// Taken from OCM@v0.4.0 because the implementation in v0.6.0 looks flawed
	if o, ok := e.(*compdesc.Resource); !ok {
		return false
	} else {
		if !reflect.DeepEqual(&r.ElementMeta, &o.ElementMeta) {
			return false
		}
		if !reflect.DeepEqual(&r.Access, &o.Access) {
			return false
		}
		return r.Type == o.Type &&
			r.Relation == o.Relation &&
			reflect.DeepEqual(r.SourceRef, o.SourceRef)
	}
}
