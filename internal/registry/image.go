package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/registry/portforward"
	"k8s.io/apimachinery/pkg/util/httpstream"
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

// for testing
type utils struct {
	daemonImage        func(name.Reference, ...daemon.Option) (v1.Image, error)
	portforwardNewDial func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error)
	remoteWrite        func(ref name.Reference, img v1.Image, options ...remote.Option) error
}

func ImportImage(ctx context.Context, imageName string, opts ImportOptions) (string, clierror.Error) {
	return importImage(ctx, imageName, opts, utils{
		daemonImage:        daemon.Image,
		portforwardNewDial: portforward.NewDialFor,
		remoteWrite:        remote.Write,
	})
}

func importImage(ctx context.Context, imageName string, opts ImportOptions, utils utils) (string, clierror.Error) {
	localImage, err := imageFromInternalRegistry(ctx, imageName, utils)
	if err != nil {
		return "", clierror.Wrap(err,
			clierror.New("failed to load image from local docker daemon",
				"make sure docker daemon is running",
				"make sure the image exists in the local docker daemon",
			),
		)
	}

	conn, err := utils.portforwardNewDial(opts.ClusterAPIRestConfig, opts.RegistryPodName, opts.RegistryPodNamespace)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to create registry portforward connection"))
	}
	defer conn.Close()

	localTr := portforward.NewPortforwardTransport(conn, opts.RegistryPodPort)
	transport := portforward.NewOnErrRetryTransport(localTr)

	pushedImage, err := imageToInClusterRegistry(ctx, localImage, transport, opts.RegistryAuth, opts.RegistryPullHost, imageName, utils)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to push image to the in-cluster registry"))
	}

	return pushedImage, nil
}

func imageFromInternalRegistry(ctx context.Context, userImage string, utils utils) (v1.Image, error) {
	tag, err := name.NewTag(userImage, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	// check if user defined custom registry - what is not allowed
	if tag.RegistryStr() != name.DefaultRegistry {
		return nil, fmt.Errorf("image '%s' can't contain registry '%s' address", tag.String(), tag.RegistryStr())
	}

	return utils.daemonImage(tag, daemon.WithContext(ctx))
}

func imageToInClusterRegistry(ctx context.Context, image v1.Image, transport http.RoundTripper, auth authn.Authenticator, pullHost, userImageName string, utils utils) (string, error) {
	tag, err := name.NewTag(userImageName, name.WeakValidation)
	if err != nil {
		return "", err
	}

	newReg, err := name.NewRegistry(pullHost, name.WeakValidation, name.Insecure)
	if err != nil {
		return "", err
	}
	tag.Registry = newReg

	progress := make(chan v1.Update, 100)

	go utils.remoteWrite(tag, image,
		remote.WithTransport(transport),
		remote.WithAuth(auth),
		remote.WithContext(ctx),
		remote.WithProgress(progress),
	)

	for u := range progress {
		switch {
		case u.Error != nil && errors.Is(u.Error, io.EOF):
			return fmt.Sprintf("%s/%s:%s", tag.RegistryStr(), tag.RepositoryStr(), tag.TagStr()), nil
		case u.Error != nil:
			return "", fmt.Errorf("error pushing image: %w", u.Error)
		default:
			fmt.Printf("pushing image is in progress: %d/%d\n", u.Complete, u.Total)
		}
	}

	return fmt.Sprintf("%s/%s:%s", tag.RegistryStr(), tag.RepositoryStr(), tag.TagStr()), nil
}
