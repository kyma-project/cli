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
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/portforward"
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

func ImportImage(ctx context.Context, imageName string, pushFunc PushFunc) (string, clierror.Error) {
	return importImage(ctx, imageName, pushFunc, utils{
		daemonImage:        daemon.Image,
		portforwardNewDial: portforward.NewDialFor,
		remoteWrite:        remote.Write,
	})
}

func importImage(ctx context.Context, imageName string, pushFunc PushFunc, utils utils) (string, clierror.Error) {
	localImage, err := imageFromInternalRegistry(ctx, imageName, utils)
	if err != nil {
		return "", clierror.Wrap(err,
			clierror.New("failed to load the image from the local Docker daemon",
				"ensure the Docker daemon is running",
				"ensure the image exists in the local Docker daemon",
			),
		)
	}

	return pushFunc(ctx, imageName, localImage, utils)
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

type PushFunc func(context.Context, string, v1.Image, utils) (string, clierror.Error)

func NewPushFunc(registryAddress string, registryAuth authn.Authenticator) PushFunc {
	return func(ctx context.Context, imageName string, localImage v1.Image, utils utils) (string, clierror.Error) {
		pushedImage, err := imageToInClusterRegistry(ctx, localImage, remote.DefaultTransport, registryAuth, registryAddress, imageName, utils)
		if err != nil {
			return "", clierror.Wrap(err, clierror.New("failed to push image to the in-cluster registry"))
		}

		return pushedImage, nil
	}
}

func NewPushWithPortforwardFunc(clusterAPIRestConfig *rest.Config, registryPodName, registryPodNamespace, registryPodPort, registryPullHost string, registryAuth authn.Authenticator) PushFunc {
	return func(ctx context.Context, imageName string, localImage v1.Image, utils utils) (string, clierror.Error) {
		conn, err := utils.portforwardNewDial(clusterAPIRestConfig, registryPodName, registryPodNamespace)
		if err != nil {
			return "", clierror.Wrap(err, clierror.New("failed to create registry portforward connection"))
		}
		defer conn.Close()

		localTr := portforward.NewPortforwardTransport(conn, registryPodPort)
		transport := portforward.NewOnErrRetryTransport(localTr)

		pushedImage, err := imageToInClusterRegistry(ctx, localImage, transport, registryAuth, registryPullHost, imageName, utils)
		if err != nil {
			return "", clierror.Wrap(err, clierror.New("failed to push image to the in-cluster registry", "pushing through portforward may be unstable, expose the registry in the Registry CR"))
		}

		return pushedImage, nil
	}
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

	err = utils.remoteWrite(tag, image,
		remote.WithTransport(transport),
		remote.WithAuth(auth),
		remote.WithContext(ctx),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", tag.RepositoryStr(), tag.TagStr()), nil
}
