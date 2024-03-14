package registry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/cli.v3/internal/registry/portforward"
	"k8s.io/client-go/rest"
)

type ImportOptions struct {
	ClusterAPIRestConfig *rest.Config
	RegistryAuth         authn.Authenticator
	RegistryPullHost     string
	RegistryPodName      string
	RegistryPodNamespace string
	RegistryPodPort      string
}

func ImportImage(ctx context.Context, imageName string, opts ImportOptions) (string, error) {
	localImage, err := imageFromInternalRegistry(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("failed to load image from local docker daemon: %s", err.Error())
	}

	conn, err := portforward.NewDialFor(opts.ClusterAPIRestConfig, opts.RegistryPodName, opts.RegistryPodNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to create registry portforward connection: %s", err.Error())
	}
	defer conn.Close()

	localTr := portforward.NewPortforwardTransport(conn, opts.RegistryPodPort)
	transport := portforward.NewRetryTransport(localTr)

	pushedImage, err := imageToInClusterRegistry(ctx, localImage, transport, opts.RegistryAuth, opts.RegistryPullHost, imageName)
	if err != nil {
		return "", fmt.Errorf("failed to push image to the in-cluster registry: %s", err.Error())
	}

	return pushedImage, nil
}

func imageFromInternalRegistry(ctx context.Context, userImage string) (v1.Image, error) {
	tag, err := name.NewTag(userImage, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	if tag.RegistryStr() != name.DefaultRegistry {
		return nil, fmt.Errorf("image '%s' can't contain registry '%s' address", tag.String(), tag.RegistryStr())
	}

	return daemon.Image(tag, daemon.WithContext(ctx))
}

func imageToInClusterRegistry(ctx context.Context, image v1.Image, transport http.RoundTripper, auth authn.Authenticator, pullHost, userImageName string) (string, error) {
	tag, err := name.NewTag(userImageName, name.WeakValidation)
	if err != nil {
		return "", err
	}

	newReg, err := name.NewRegistry(pullHost, name.WeakValidation, name.Insecure)
	if err != nil {
		return "", err
	}
	tag.Registry = newReg

	err = remote.Write(tag, image,
		remote.WithTransport(transport),
		remote.WithAuth(auth),
		remote.WithContext(ctx),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s:%s", tag.RegistryStr(), tag.RepositoryStr(), tag.TagStr()), nil
}
