package module

import (
	"fmt"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2Sign "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	"github.com/gardener/component-spec/bindings-go/ctf"
)

func Sign() error {
	signer, err := cdv2Sign.CreateRSASignerFromKeyFile(c.opts.PrivateKeyPath, cdv2.MediaTypePEM)
	if err != nil {
		c.CurrentStep.Failure()
		return fmt.Errorf("unable to create rsa signer: %w", err)
	}
	digestedCds, err := recursivelyAddDigestsToCd(cd, *repoCtx, ociClient, blobResolvers, context.TODO(), skipAccessTypesMap)
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

func recursivelyAddDigestsToCd(cd *cdv2.ComponentDescriptor, repoContext cdv2.OCIRegistryRepository, ociClient ociclient.Client, blobResolvers map[string]ctf.BlobResolver, ctx context.Context, skipAccessTypes map[string]bool) ([]*cdv2.ComponentDescriptor, error) {
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

		cds, err := recursivelyAddDigestsToCd(childCd, repoContext, ociClient, blobResolvers, ctx, skipAccessTypes)
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

	// set the do not sign digest notation on skip-access-type resources
	for i, res := range cd.Resources {
		res := res
		if _, ok := skipAccessTypes[res.Access.Type]; ok {
			log := logger.Log.WithValues("componentDescriptor", cd, "resource.name", res.Name, "resource.version", res.Version, "resource.extraIdentity", res.ExtraIdentity)
			log.Info(fmt.Sprintf("adding %s digest to resource based on skip-access-type", cdv2.ExcludeFromSignature))

			res.Digest = cdv2.NewExcludeFromSignatureDigest()
			cd.Resources[i] = res
		}
	}

	digester := NewDigester(ociClient, *hasher)
	if err := cdv2Sign.AddDigestsToComponentDescriptor(context.TODO(), cd, cdResolver, digester.DigestForResource); err != nil {
		return nil, fmt.Errorf("failed adding digests to cd %s:%s: %w", cd.Name, cd.Version, err)
	}
	cdsWithHashes = append(cdsWithHashes, cd)
	return cdsWithHashes, nil
}
