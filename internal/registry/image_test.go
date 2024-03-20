package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/cli.v3/internal/registry/portforward/automock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
)

func Test_importImage(t *testing.T) {
	type args struct {
		ctx       context.Context
		imageName string
		opts      ImportOptions
		utils     utils
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "import image",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				opts: ImportOptions{
					ClusterAPIRestConfig: nil,
					RegistryAuth: &basicAuth{
						username: "username",
						password: "password",
					},
					RegistryPullHost:     "testhost:123",
					RegistryPodName:      "podname",
					RegistryPodNamespace: "podnamespace",
					RegistryPodPort:      "1234",
				},
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						require.Equal(t, "index.docker.io/library/test:image", r.Name())
						require.Len(t, o, 1)

						return &fake.FakeImage{}, nil
					},
					portforwardNewDial: func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
						require.Nil(t, config)
						require.Equal(t, "podname", podName)
						require.Equal(t, "podnamespace", podNamespace)

						mock := automock.NewConnection(t)
						mock.On("Close").Return(nil).Once()
						return mock, nil
					},
					remoteWrite: func(ref name.Reference, img v1.Image, o ...remote.Option) error {
						require.Equal(t, "testhost:123/test:image", ref.Name())
						require.Equal(t, &fake.FakeImage{}, img)
						require.Len(t, o, 3)

						return nil
					},
				},
			},
			wantErr: nil,
			want:    "testhost:123/test:image",
		},
		{
			name: "wrong image format error",
			args: args{
				imageName: ":::::::::",
			},
			wantErr: errors.New("failed to load image from local docker daemon: repository can only contain the characters `abcdefghijklmnopqrstuvwxyz0123456789_-./`: ::::::::"),
		},
		{
			name: "image contains registry address error",
			args: args{
				imageName: "gcr.io/test:image",
			},
			wantErr: errors.New("failed to load image from local docker daemon: image 'gcr.io/test:image' can't contain registry 'gcr.io' address"),
		},
		{
			name: "get image from local daemon error",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						return nil, errors.New("test-error")
					},
				},
			},
			wantErr: errors.New("failed to load image from local docker daemon: test-error"),
		},
		{
			name: "create new portforward dial error",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						return &fake.FakeImage{}, nil
					},
					portforwardNewDial: func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
						return nil, errors.New("test-error")
					},
				},
			},
			wantErr: errors.New("failed to create registry portforward connection: test-error"),
		},
		{
			name: "wrong PullHost format",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				opts: ImportOptions{
					RegistryPullHost: "<    >",
				},
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						return &fake.FakeImage{}, nil
					},
					portforwardNewDial: func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
						mock := automock.NewConnection(t)
						mock.On("Close").Return(nil).Once()
						return mock, nil
					},
				},
			},
			wantErr: errors.New("failed to push image to the in-cluster registry: registries must be valid RFC 3986 URI authorities: <    >"),
		},
		{
			name: "import image",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				opts: ImportOptions{
					ClusterAPIRestConfig: nil,
					RegistryAuth: &basicAuth{
						username: "username",
						password: "password",
					},
					RegistryPullHost:     "testhost:123",
					RegistryPodName:      "podname",
					RegistryPodNamespace: "podnamespace",
					RegistryPodPort:      "1234",
				},
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						return &fake.FakeImage{}, nil
					},
					portforwardNewDial: func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
						mock := automock.NewConnection(t)
						mock.On("Close").Return(nil).Once()
						return mock, nil
					},
					remoteWrite: func(ref name.Reference, img v1.Image, o ...remote.Option) error {
						return errors.New("test error")
					},
				},
			},
			wantErr: errors.New("failed to push image to the in-cluster registry: test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := importImage(tt.args.ctx, tt.args.imageName, tt.args.opts, tt.args.utils)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}
