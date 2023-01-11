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

// Remote represents remote OCI registry and the means to access it
type Remote struct {
	Registry    string
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
		nameMapping := cdv2.OCIRegistryURLPathMapping
		if rc := archive.ComponentDescriptor.GetEffectiveRepositoryContext(); rc != nil {
			var repo cdv2.OCIRegistryRepository
			if err = rc.DecodeInto(&repo); err != nil {
				return errors.Wrap(err, "unable to decode component descriptor")
			}

			if repo.ComponentNameMapping == cdv2.OCIRegistryDigestMapping {
				nameMapping = cdv2.OCIRegistryDigestMapping
			}
		}
		if err := cdv2.InjectRepositoryContext(archive.ComponentDescriptor, cdv2.NewOCIRegistryRepository(r.Registry, nameMapping)); err != nil {
			return errors.Wrap(err, "unable to add repository context to component descriptor")
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

func Pull() {
	// TODO
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
