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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type NameMapping ocireg.ComponentNameMapping

const (
	URLPathNameMapping            = NameMapping(ocireg.OCIRegistryURLPathMapping)
	DigestNameMapping             = NameMapping(ocireg.OCIRegistryDigestMapping)
	accessLocalReferenceFieldName = "localReference"
)

// Remote represents remote OCI registry and the means to access it
type Remote struct {
	Registry    string
	NameMapping NameMapping
	Credentials string
	Token       string
	Insecure    bool
	OciRepoAccess
}

func (r *Remote) GetRepository(ctx cpi.Context) (cpi.Repository, error) {
	creds := r.getCredentials(ctx)
	repoType := oci.Type

	url := NoSchemeURL(r.Registry)
	if r.Insecure {
		url = fmt.Sprintf("http://%s", url)
	}

	ociRepoSpec := &oci.RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(repoType),
		BaseURL:             url,
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

func (r *Remote) getCredentials(ctx cpi.Context) credentials.Credentials {
	if r.Insecure {
		return credentials.NewCredentials(nil)
	}
	var creds credentials.Credentials
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".docker", "config.json")
		if repo, err := dockerconfig.NewRepository(ctx.CredentialsContext(), path, nil, true); err == nil {
			// this uses the first part of the url to resolve the correct host, e.g.
			// ghcr.io/jakobmoellersap/testmodule => ghcr.io
			hostNameInDockerConfigJSON := strings.Split(NoSchemeURL(r.Registry), "/")[0]
			if creds, err = repo.LookupCredentials(hostNameInDockerConfigJSON); err != nil {
				// this forces creds to be nil in case the host was not found in the native docker store
				creds = nil
			}
		}
	}

	// if no creds are set, try to use username and password that are provided.
	if creds == nil || isEmptyAuth(creds) {
		u, p := r.userPass()
		if p == "" {
			p = r.Token
		}
		creds = credentials.DirectCredentials{
			"username": u,
			"password": p,
		}
	}
	return creds
}

func isEmptyAuth(creds credentials.Credentials) bool {
	if len(creds.GetProperty("auth")) != 0 {
		return false
	}
	if len(creds.GetProperty("username")) != 0 {
		return false
	}

	return true
}

// userPass splits the credentials string into user and password.
// If the string is empty or can't be split, it returns 2 empty strings.
func (r *Remote) userPass() (string, string) {
	u, p, found := strings.Cut(r.Credentials, ":")
	if !found {
		return "", ""
	}
	return u, p
}

func NoSchemeURL(url string) string {
	regex := regexp.MustCompile(`^https?://`)
	return regex.ReplaceAllString(url, "")
}

func (r *Remote) ShouldPushArchive(repo cpi.Repository, archive *comparch.ComponentArchive, overwrite bool) (bool,
	error) {
	if !overwrite {
		versionExists, err := r.ComponentVersionExists(archive, repo)
		if err != nil {
			fmt.Println("========================================")
			fmt.Println(fmt.Errorf("could not check if the component version exists: %w", err))
			fmt.Println("========================================")
		}

		if versionExists {
			versionAccess, err := r.GetComponentVersion(archive, repo)
			if err != nil {
				return false, fmt.Errorf("could not lookup component version: %w", err)
			}

			if r.DescriptorResourcesAreEquivalent(archive, versionAccess) {
				return false, nil
			}
			return false, fmt.Errorf("version %s already exists with different content, please use "+
				"--module-archive-version-overwrite flag to overwrite it",
				archive.GetVersion())
		}
	}

	return true, nil
}

// Push picks up the archive described in the config and pushes it to the provided registry.
// The credentials and token are optional parameters
func (r *Remote) Push(repo cpi.Repository, archive *comparch.ComponentArchive,
	overwrite bool) (ocm.ComponentVersionAccess, error) {
	if err := r.PushComponentVersion(archive, repo, overwrite); err != nil {
		return nil, err
	}

	componentVersion, err := r.GetComponentVersion(archive, repo)

	return componentVersion, err
}

type customTransferHandler struct {
	transferhandler.TransferHandler
}

func (h *customTransferHandler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess,
	meta *compdesc.ComponentReference, tgt ocm.Repository) (ocm.ComponentVersionAccess, transferhandler.TransferHandler,
	error) {
	return h.TransferHandler.TransferVersion(repo, src, meta, tgt)
}
