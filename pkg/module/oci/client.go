package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"go.uber.org/zap"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/opencontainers/go-digest"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultTimeout = 10 * time.Second
)

type Client interface {
	// GetRawManifest returns the raw manifest for a reference.
	// The returned manifest can either be single arch or multi arch (image index/manifest list)
	GetRawManifest(ctx context.Context, ref string) (ocispecv1.Descriptor, []byte, error)

	// GetManifest returns the ocispec manifest for a reference
	GetManifest(ctx context.Context, ref string) (*ocispecv1.Manifest, error)

	// Fetch fetches the blob for the given ocispec Descriptor
	Fetch(ctx context.Context, ref string, desc ocispecv1.Descriptor, writer io.Writer) error

	// PushManifest uploads the given manifest with all its layers to the given reference in the registry configured in the client.
	PushManifest(ctx context.Context, ref string, manifest *artdesc.Manifest) error

	// Cache exposes the client's cache where all intermediate manifests and data are stored.
	Cache() Cache
}

type client struct {
	cache    Cache
	registry string
	logger   *zap.SugaredLogger
	// user to authenticate when calling the registry configured in the client
	user string
	// secret can be either a password (if user provided) or a long-lived token.
	secret string
	// timeout for all network calls the client will perform
	timeout time.Duration
	// if true, the client will make all calls with http instead of https
	insecure bool
}

type Options struct {
	Registry string
	// (Optional) user to authenticate when calling the registry configured in the client
	User string
	// (Optional) secret can be either a password (if user provided) or a long-lived token.
	Secret string
	// timeout for all network calls the client will perform
	Timeout time.Duration
	// if true, the client will make all calls with http instead of https
	Insecure bool
}

func NewClient(o *Options, logger *zap.SugaredLogger) (Client, error) {
	c := &client{
		registry: o.Registry,
		user:     o.User,
		secret:   o.Secret,
		timeout:  o.Timeout,
		insecure: o.Insecure,
		logger:   logger,
	}

	if c.timeout == 0 {
		c.timeout = DefaultTimeout
	}

	c.cache = NewInMemoryCache()

	return c, nil
}

func (c *client) Cache() Cache {
	return c.cache
}

func (c *client) GetRawManifest(ctx context.Context, ref string) (ocispecv1.Descriptor, []byte, error) {
	refspec, err := ParseRef(ref)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to parse ref: %w", err)
	}
	ref = refspec.String()

	resolver := c.resolver()

	_, desc, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, err
	}

	if desc.MediaType == MediaTypeDockerV2Schema1Manifest || desc.MediaType == MediaTypeDockerV2Schema1SignedManifest {
		convertedManifestDesc, err := ConvertV1ManifestToV2(ctx, c, c.cache, ref, desc)
		if err != nil {
			return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to convert v1 manifest to v2: %w", err)
		}
		desc = convertedManifestDesc
	}

	if !isSingleArchImage(desc.MediaType) && !isMultiArchImage(desc.MediaType) {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf(
			"media type is not an image manifest or image index: %s", desc.MediaType,
		)
	}

	data := bytes.NewBuffer([]byte{})
	if err := c.Fetch(ctx, ref, desc, data); err != nil {
		return ocispecv1.Descriptor{}, nil, err
	}
	rawManifest := data.Bytes()

	return desc, rawManifest, nil
}

func (c *client) GetManifest(ctx context.Context, ref string) (*ocispecv1.Manifest, error) {
	desc, rawManifest, err := c.GetRawManifest(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("unable to get manifest: %w", err)
	}

	if desc.MediaType != ocispecv1.MediaTypeImageManifest && desc.MediaType != images.MediaTypeDockerSchema2Manifest {
		return nil, fmt.Errorf("media type is not an image manifest: %s", desc.MediaType)
	}

	var manifest ocispecv1.Manifest
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return nil, fmt.Errorf("unable to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

func (c *client) Fetch(ctx context.Context, ref string, desc ocispecv1.Descriptor, writer io.Writer) error {
	refspec, err := ParseRef(ref)
	if err != nil {
		return fmt.Errorf("unable to parse ref: %w", err)
	}
	ref = refspec.String()

	reader, err := c.getFetchReader(ctx, ref, desc)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			c.logger.Error(err, "failed closing reader", "ref", ref)
		}
	}()

	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}
	return nil
}

func (c *client) getFetchReader(ctx context.Context, ref string, desc ocispecv1.Descriptor) (io.ReadCloser, error) {
	if c.cache != nil {
		reader, err := c.cache.Get(desc)
		if err != nil && !errors.Is(ErrNotFound, err) {
			return nil, err
		}
		if err == nil {
			return reader, nil
		}
	}

	resolver := c.resolver()

	fetcher, err := resolver.Fetcher(ctx, ref)
	if err != nil {
		return nil, err
	}
	reader, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	// try to cache
	if c.cache != nil {
		if err := c.cache.Add(desc, reader); err != nil {
			// do not throw an error as cache is just an optimization
			if err = reader.Close(); err != nil {
				c.logger.Info("unable to close reader", "ref", ref, "error", err.Error())

			}
			return fetcher.Fetch(ctx, desc)
		}
		return c.cache.Get(desc)
	}

	return reader, err
}

func (c *client) PushManifest(ctx context.Context, ref string, manifest *artdesc.Manifest) error {
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("unable to marshal manifest: %w", err)
	}

	desc := ocispecv1.Descriptor{
		MediaType:   ocispecv1.MediaTypeImageManifest,
		Digest:      digest.FromBytes(manifestBytes),
		Size:        int64(len(manifestBytes)),
		Annotations: manifest.Annotations,
	}

	resolver := c.resolver()

	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return err
	}

	if isSingleArchImage(desc.MediaType) {
		manifest := ocispecv1.Manifest{}
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return fmt.Errorf("unable to unmarshal manifest: %w", err)
		}

		// add dummy config if it is not set
		if manifest.Config.Size == 0 {
			dummyConfig := []byte("{}")
			manifest.Config = ocispecv1.Descriptor{
				MediaType: "application/json",
				Digest:    digest.FromBytes(dummyConfig),
				Size:      int64(len(dummyConfig)),
			}
			if err := c.cache.Add(manifest.Config, io.NopCloser(bytes.NewBuffer(dummyConfig))); err != nil {
				return fmt.Errorf("unable to add dummy config to cache: %w", err)
			}
		}

		if err := c.pushContent(ctx, c.cache, pusher, manifest.Config); err != nil {
			return fmt.Errorf("unable to push config: %w", err)

		}

		for _, layerDesc := range manifest.Layers {
			if err := c.pushContent(ctx, c.cache, pusher, layerDesc); err != nil {
				return fmt.Errorf("unable to push layer: %w", err)
			}
		}
	}

	if err := c.cache.Add(desc, io.NopCloser(bytes.NewBuffer(manifestBytes))); err != nil {
		return fmt.Errorf("unable to add manifest to cache: %w", err)
	}

	if err := c.pushContent(ctx, c.cache, pusher, desc); err != nil {
		return fmt.Errorf("unable to push manifest: %w", err)
	}

	return nil
}

func (c *client) pushContent(ctx context.Context, store Store, pusher remotes.Pusher, desc ocispecv1.Descriptor) error {
	if store == nil {
		return errors.New("you must define a store to upload content")
	}
	r, err := store.Get(desc)
	if err != nil {
		return err
	}
	defer r.Close()

	writer, err := pusher.Push(addKnownMediaTypesToCtx(ctx, []string{desc.MediaType}), desc)
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	defer writer.Close()
	return content.Copy(ctx, writer, r, desc.Size, desc.Digest)
}

// AddKnownMediaTypesToCtx adds a list of known media types to the context
func addKnownMediaTypesToCtx(ctx context.Context, mediaTypes []string) context.Context {
	for _, mediaType := range mediaTypes {
		ctx = remotes.WithMediaTypeKeyPrefix(ctx, mediaType, "custom")
	}
	return ctx
}

func isSingleArchImage(mediaType string) bool {
	return mediaType == ocispecv1.MediaTypeImageManifest ||
		mediaType == images.MediaTypeDockerSchema2Manifest
}

func isMultiArchImage(mediaType string) bool {
	return mediaType == ocispecv1.MediaTypeImageIndex ||
		mediaType == images.MediaTypeDockerSchema2ManifestList
}

// resolver returns an authenticated remote resolver for a reference.
func (c *client) resolver() remotes.Resolver {
	scheme := "https"
	if c.insecure {
		scheme = "http"
	}
	do := docker.ResolverOptions{
		Hosts: func(host string) ([]docker.RegistryHost, error) {
			config := docker.RegistryHost{
				Client:       &http.Client{Timeout: c.timeout},
				Host:         host,
				Scheme:       scheme,
				Path:         "/v2",
				Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
			}

			//The secret alone may be enough for token-based auth scenarios
			if c.secret != "" {
				f := func(host string) (string, string, error) {
					//We do nothing with the host value
					return c.user, c.secret, nil
				}

				config.Authorizer = docker.NewDockerAuthorizer(docker.WithAuthCreds(f))

			}
			return []docker.RegistryHost{config}, nil
		},
	}

	return docker.NewResolver(do)
}
