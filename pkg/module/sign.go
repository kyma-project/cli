package module

import (
	"context"
	"fmt"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2Sign "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/kyma-project/cli/pkg/module/signatures"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ComponentSignConfig struct {
	Name           string // Name of the module (mandatory)
	Version        string // Version of the module (mandatory)
	PrivateKeyPath string // The private key used for signing (mandatory)
	SignatureName  string // Name of the signature for signing
}

func Sign(cfg *ComponentSignConfig, remote *Remote, log *zap.SugaredLogger) ([]*cdv2.ComponentDescriptor, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	repoCtx := BuildNewOCIRegistryRepository(remote.Registry, cdv2.ComponentNameMapping(remote.NameMapping))

	signer, err := cdv2Sign.CreateRSASignerFromKeyFile(cfg.PrivateKeyPath, cdv2.MediaTypePEM)
	if err != nil {
		return nil, fmt.Errorf("unable to create rsa signer: %w", err)
	}

	u, p := remote.UserPass()
	ociClient, err := oci.NewClient(&oci.Options{
		Registry: remote.Registry,
		User:     u,
		Secret:   p,
		Insecure: remote.Insecure,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("unable to build oci client: %w", err)
	}

	cdresolver := cdoci.NewResolver(ociClient)
	cd, blobResolver, err := cdresolver.ResolveWithBlobResolver(ctx, repoCtx, cfg.Name, cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to to fetch component descriptor %s:%s: %w", cfg.Name, cfg.Version, err)
	}

	blobResolvers := map[string]ctf.BlobResolver{}
	blobResolvers[fmt.Sprintf("%s:%s", cd.Name, cd.Version)] = blobResolver

	digestedCds, err := recursivelyAddDigestsToCd(ctx, cd, *repoCtx, ociClient, blobResolvers)
	if err != nil {
		return nil, fmt.Errorf("unable to add digests to component descriptor: %w", err)
	}

	for _, digestedCd := range digestedCds {
		hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
		if err != nil {
			return nil, fmt.Errorf("unable to create hasher: %w", err)
		}

		if err := cdv2Sign.SignComponentDescriptor(digestedCd, signer, *hasher, cfg.SignatureName); err != nil {
			return nil, fmt.Errorf("unable to sign component descriptor: %w", err)
		}
		log.Info(fmt.Sprintf("Signed component descriptor %s %s", digestedCd.Name, digestedCd.Version))
	}

	return digestedCds, nil
}

func recursivelyAddDigestsToCd(ctx context.Context, cd *cdv2.ComponentDescriptor, repoContext cdv2.OCIRegistryRepository, ociClient oci.Client, blobResolvers map[string]ctf.BlobResolver) ([]*cdv2.ComponentDescriptor, error) {
	cdsWithHashes := []*cdv2.ComponentDescriptor{}

	cdResolver := func(c context.Context, cd cdv2.ComponentDescriptor, cr cdv2.ComponentReference) (*cdv2.DigestSpec, error) {
		ociRef, err := cdoci.OCIRef(repoContext, cr.Name, cr.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid component reference: %w", err)
		}

		cdresolver := cdoci.NewResolver(ociClient)
		childCd, blobResolver, err := cdresolver.ResolveWithBlobResolver(ctx, &repoContext, cr.ComponentName, cr.Version)
		if err != nil {
			return nil, fmt.Errorf("unable to to fetch component descriptor %s: %w", ociRef, err)
		}
		blobResolvers[fmt.Sprintf("%s:%s", childCd.Name, childCd.Version)] = blobResolver

		cds, err := recursivelyAddDigestsToCd(ctx, childCd, repoContext, ociClient, blobResolvers)
		if err != nil {
			return nil, fmt.Errorf("failed resolving referenced cd %s:%s: %w", cr.Name, cr.Version, err)
		}
		cdsWithHashes = append(cdsWithHashes, cds...)

		hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed creating hasher: %w", err)
		}
		hashCd, err := cdv2Sign.HashForComponentDescriptor(*childCd, *hasher)
		if err != nil {
			return nil, fmt.Errorf("failed hashing referenced cd %s:%s: %w", cr.Name, cr.Version, err)
		}
		return hashCd, nil
	}

	hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed creating hasher: %w", err)
	}

	digester := signatures.NewDigester(ociClient, *hasher)
	if err := cdv2Sign.AddDigestsToComponentDescriptor(ctx, cd, cdResolver, digester.DigestForResource); err != nil {
		return nil, fmt.Errorf("failed adding digests to cd %s:%s: %w", cd.Name, cd.Version, err)
	}
	cdsWithHashes = append(cdsWithHashes, cd)
	return cdsWithHashes, nil
}

func (cfg *ComponentSignConfig) validate() error {
	if cfg.Name == "" {
		return errors.New("The module name cannot be empty")
	}
	if cfg.Version == "" {
		return errors.New("The module version cannot be empty")
	}
	if cfg.PrivateKeyPath == "" {
		return errors.New("The private key path cannot be empty")
	}
	if cfg.SignatureName == "" {
		return errors.New("The signature name cannot be empty")
	}

	return nil
}
