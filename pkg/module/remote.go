package module

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	oci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	componentTransfer "github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
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
		if home, err := os.UserHomeDir(); err == nil {
			path := filepath.Join(home, ".docker", "config.json")
			if repo, err := dockerconfig.NewRepository(ctx.CredentialsContext(), path, true); err == nil {
				// this uses the first part of the url to resolve the correct host, e.g.
				// ghcr.io/jakobmoellersap/testmodule => ghcr.io
				hostNameInDockerConfigJson := strings.Split(NoSchemeURL(r.Registry), "/")[0]
				creds, err = repo.LookupCredentials(hostNameInDockerConfigJson)
				if err != nil {
					// this forces creds to be nil in case the host was not found in the native docker store
					creds = nil
				}
			}
		}
		// if no creds are set, try to use username and password that are provided.
		if creds == nil {
			u, p := r.UserPass()
			creds = credentials.DirectCredentials{
				"username": u,
				"password": p,
			}
		}
	}
	var repoType string
	if compatattr.Get(ctx) {
		repoType = oci.LegacyType
	} else {
		repoType = oci.Type
	}

	ociRepoSpec := &oci.RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(repoType),
		BaseURL:             NoSchemeURL(r.Registry),
	}
	genericSpec := genericocireg.NewRepositorySpec(
		ociRepoSpec, &ocireg.ComponentRepositoryMeta{
			ComponentNameMapping: ocireg.ComponentNameMapping(r.NameMapping),
		},
	)

	repo, err := ctx.RepositoryForSpec(genericSpec, creds)

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
func Push(archive *comparch.ComponentArchive, r *Remote, overwrite bool) (ocm.ComponentVersionAccess, error) {
	repo, err := r.GetRepository(archive.GetContext())
	if err != nil {
		return nil, err
	}

	transferHandler, err := standard.New(standard.Overwrite(overwrite))
	if err != nil {
		return nil, fmt.Errorf("could not setup archive transfer: %w", err)
	}

	if err = componentTransfer.TransferVersion(
		nil, nil, archive.ComponentVersionAccess, repo, transferHandler,
	); err != nil {
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
