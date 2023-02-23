package module

import (
	"context"
	"strings"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NameMapping cdv2.ComponentNameMapping

const (
	URLPathNameMapping = NameMapping(cdv2.OCIRegistryURLPathMapping)
	DigestNameMapping  = NameMapping(cdv2.OCIRegistryDigestMapping)
)

// Remote represents remote OCI registry and the means to access it
type Remote struct {
	Registry    string
	NameMapping NameMapping
	Credentials string
	Token       string
	Insecure    bool
}

// Push picks up the archive described in the config and pushes it to the provided registry.
// The credentials and token are optional parameters
func Push(archive *ctf.ComponentArchive, r *Remote, log *zap.SugaredLogger) error {

	u, p := r.UserPass()
	ociClient, err := oci.NewClient(&oci.Options{
		Registry: r.Registry,
		User:     u,
		Secret:   p,
		Insecure: r.Insecure,
	}, log)

	if err != nil {
		return errors.Wrap(err, "unable to create an OCI client")
	}
	ctx := context.Background()

	// update repository context
	if len(r.Registry) != 0 {
		if rc := archive.ComponentDescriptor.GetEffectiveRepositoryContext(); rc != nil {
			//This code executes, for example, during push of the existing module (repo 1) to another repository (repo 2). A valid scenario for the CLI "sign module" cmd.
			var repo cdv2.OCIRegistryRepository
			if err = rc.DecodeInto(&repo); err != nil {
				return errors.Wrap(err, "unable to decode component descriptor")
			}

			//Inject only if the repo is different
			if repo.BaseURL != NoSchemeURL(r.Registry) || repo.ComponentNameMapping != cdv2.ComponentNameMapping(r.NameMapping) {
				if err := cdv2.InjectRepositoryContext(archive.ComponentDescriptor, BuildNewOCIRegistryRepository(r.Registry, cdv2.ComponentNameMapping(r.NameMapping))); err != nil {
					return errors.Wrap(err, "unable to add repository context to component descriptor")
				}
			}
		}
	}

	manifest, err := cdoci.NewManifestBuilder(ociClient.Cache(), archive).Build(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to build OCI artifact for component archive")
	}

	ref, err := oci.Ref(archive.ComponentDescriptor.GetEffectiveRepositoryContext(), archive.ComponentDescriptor.Name, archive.ComponentDescriptor.Version)
	if err != nil {
		return errors.Wrap(err, "invalid component reference")
	}
	if err := ociClient.PushManifest(ctx, ref, manifest); err != nil {
		return err
	}
	log.Debugf("Successfully uploaded manifest at %q", ref)

	return nil
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
