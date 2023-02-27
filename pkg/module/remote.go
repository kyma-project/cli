package module

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	oci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	componentTransfer "github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	ocmerrors "github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type NameMapping ocireg.ComponentNameMapping

const (
	URLPathNameMapping = NameMapping(ocireg.OCIRegistryURLPathMapping)
	DigestNameMapping  = NameMapping(ocireg.OCIRegistryDigestMapping)
)

// Remote represents remote OCI registry and the means to access it
type Remote struct {
	Registry    string
	NameMapping NameMapping
	Credentials string
	Token       string
	Insecure    bool
}

func (r *Remote) GetRepository(ctx cpi.Context) (cpi.Repository, error) {
	var creds credentials.Credentials
	if !r.Insecure {
		u, p := r.UserPass()
		creds = credentials.DirectCredentials{
			"username": u,
			"password": p,
		}
	}
	var repoType string
	if compatattr.Get(ctx) {
		repoType = oci.LegacyType
	} else {
		repoType = oci.Type
	}
	repo, err := ctx.RepositoryForSpec(
		genericocireg.NewRepositorySpec(
			&oci.RepositorySpec{
				ObjectVersionedType: runtime.NewVersionedObjectType(repoType),
				BaseURL:             NoSchemeURL(r.Registry),
			}, &ocireg.ComponentRepositoryMeta{
				ComponentNameMapping: ocireg.ComponentNameMapping(r.NameMapping),
			},
		), creds,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating repository from spec: %w", err)
	}
	return repo, nil
}

func NoSchemeURL(url string) string {
	regex := regexp.MustCompile(`^https?://`)
	return regex.ReplaceAllString(url, "")
}

// Push picks up the archive described in the config and pushes it to the provided registry.
// The credentials and token are optional parameters
func Push(archive *comparch.ComponentArchive, r *Remote) (ocm.ComponentVersionAccess, error) {
	repo, err := r.GetRepository(archive.GetContext())
	if err != nil {
		return nil, err
	}

	name := archive.ComponentVersionAccess.GetName()
	version := archive.ComponentVersionAccess.GetVersion()
	exists, err := repo.ExistsComponentVersion(name, version)
	if exists {
		return nil, fmt.Errorf("component version already exists: %s, version %s", name, version)
	}
	if err != nil && !ocmerrors.IsErrNotFound(err) {
		return nil, fmt.Errorf("could not determine if component version already exists: %w", err)
	}

	err = componentTransfer.TransferVersion(
		nil, nil, archive.ComponentVersionAccess, repo, nil,
	)

	if err != nil {
		return nil, fmt.Errorf("could not finish component transfer: %w", err)
	}

	return repo.LookupComponentVersion(
		archive.ComponentVersionAccess.GetName(), archive.ComponentVersionAccess.GetVersion(),
	)
}

// UserPass splits the credentials string into user and password.
// If the string is empty or can't be split, it returns 2 empty strings.
func (r *Remote) UserPass() (string, string) {
	u, p, found := strings.Cut(r.Credentials, ":")
	if !found {
		return "", ""
	}
	return u, p
}
