package module

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	componentTransfer "github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
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

func (r *Remote) GetRepository() (cpi.Repository, error) {
	var creds credentials.Credentials
	if !r.Insecure {
		u, p := r.UserPass()
		creds = credentials.DirectCredentials{
			"username": u,
			"password": p,
		}
	}
	repo, err := cpi.DefaultContext().RepositoryForSpec(
		ocireg.NewRepositorySpec(
			NoSchemeURL(r.Registry), &ocireg.ComponentRepositoryMeta{
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

	repo, err := r.GetRepository()
	if err != nil {
		return nil, err
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
