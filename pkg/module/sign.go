package module

import (
	"context"
	"fmt"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2Sign "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"go.uber.org/zap"
)

func Sign(archive *ctf.ComponentArchive, cfg *ComponentConfig, privateKeyPath string, fs vfs.FileSystem, remote *Remote, log *zap.SugaredLogger) error {
	ctx := context.Background()
	repoCtx := cdv2.NewOCIRegistryRepository(remote.Registry, cdv2.OCIRegistryURLPathMapping)

	signer, err := cdv2Sign.CreateRSASignerFromKeyFile(privateKeyPath, cdv2.MediaTypePEM)
	if err != nil {
		return fmt.Errorf("unable to create rsa signer: %w", err)
	}

	u, p := remote.UserPass()
	ociClient, err := oci.NewClient(&oci.Options{
		Registry: remote.Registry,
		User:     u,
		Secret:   p,
		Insecure: remote.Insecure,
	})
	if err != nil {
		return fmt.Errorf("unable to build oci client: %w", err)
	}

	cdresolver := cdoci.NewResolver(ociClient)
	cd, blobResolver, err := cdresolver.ResolveWithBlobResolver(ctx, repoCtx, cfg.Name, cfg.Version)
	if err != nil {
		return fmt.Errorf("unable to to fetch component descriptor %s:%s: %w", cfg.Name, cfg.Version, err)
	}

	blobResolvers := map[string]ctf.BlobResolver{}
	blobResolvers[fmt.Sprintf("%s:%s", cd.Name, cd.Version)] = blobResolver

	digestedCds, err := recursivelyAddDigestsToCd(archive.ComponentDescriptor, *repoCtx, ociClient, blobResolvers, ctx)
	if err != nil {
		return fmt.Errorf("unable to add digests to component descriptor: %w", err)
	}

	targetRepoCtx := cdv2.NewOCIRegistryRepository(o.UploadBaseUrlForSigned, "")

	for _, digestedCd := range digestedCds {
		hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
		if err != nil {
			return fmt.Errorf("unable to create hasher: %w", err)
		}

		if err := cdv2Sign.SignComponentDescriptor(digestedCd, signer, *hasher, o.SignatureName); err != nil {
			return fmt.Errorf("unable to sign component descriptor: %w", err)
		}
		logger.Log.Info(fmt.Sprintf("Signed component descriptor %s %s", digestedCd.Name, digestedCd.Version))

		logger.Log.Info(fmt.Sprintf("Uploading to %s %s %s", o.UploadBaseUrlForSigned, digestedCd.Name, digestedCd.Version))

		if err := signatures.UploadCDPreservingLocalOciBlobs(ctx, *digestedCd, *targetRepoCtx, ociClient, cache, blobResolvers, o.Force, log); err != nil {
			return fmt.Errorf("unable to upload component descriptor: %w", err)
		}
	}
}

func recursivelyAddDigestsToCd(cd *cdv2.ComponentDescriptor, repoContext cdv2.OCIRegistryRepository, ociClient interface{}, blobResolvers map[string]ctf.BlobResolver, ctx context.Context) ([]*cdv2.ComponentDescriptor, error) {
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

		cds, err := recursivelyAddDigestsToCd(childCd, repoContext, ociClient, blobResolvers, ctx)
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

	digester := NewDigester(ociClient, *hasher)
	if err := cdv2Sign.AddDigestsToComponentDescriptor(context.TODO(), cd, cdResolver, digester.DigestForResource); err != nil {
		return nil, fmt.Errorf("failed adding digests to cd %s:%s: %w", cd.Name, cd.Version, err)
	}
	cdsWithHashes = append(cdsWithHashes, cd)
	return cdsWithHashes, nil
}
