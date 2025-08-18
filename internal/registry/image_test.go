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
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/registry/portforward/automock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
)

func Test_importImage(t *testing.T) {
	type args struct {
		ctx       context.Context
		imageName string
		pushFunc  PushFunc
		utils     utils
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr clierror.Error
	}{
		{
			name: "import image",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				pushFunc: NewPushWithPortforwardFunc(nil, "podname", "podnamespace", "1234", "testhost:123", &basicAuth{
					username: "username",
					password: "password",
				}),
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
			wantErr: clierror.Wrap(errors.New("repository can only contain the characters `abcdefghijklmnopqrstuvwxyz0123456789_-./`: ::::::::"),
				clierror.New("failed to load image from local docker daemon",
					"make sure docker daemon is running",
					"make sure the image exists in the local docker daemon"),
			),
		},
		{
			name: "image contains registry address error",
			args: args{
				imageName: "gcr.io/test:image",
			},
			wantErr: clierror.Wrap(errors.New("image 'gcr.io/test:image' can't contain registry 'gcr.io' address"),
				clierror.New("failed to load image from local docker daemon",
					"make sure docker daemon is running",
					"make sure the image exists in the local docker daemon"),
			),
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
			wantErr: clierror.Wrap(errors.New("test-error"),
				clierror.New("failed to load image from local docker daemon",
					"make sure docker daemon is running",
					"make sure the image exists in the local docker daemon"),
			),
		},
		{
			name: "create new portforward dial error",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				pushFunc:  NewPushWithPortforwardFunc(nil, "", "", "", "", nil),
				utils: utils{
					daemonImage: func(r name.Reference, o ...daemon.Option) (v1.Image, error) {
						return &fake.FakeImage{}, nil
					},
					portforwardNewDial: func(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
						return nil, errors.New("test-error")
					},
				},
			},
			wantErr: clierror.Wrap(errors.New("test-error"),
				clierror.New("failed to create registry portforward connection"),
			),
		},
		{
			name: "wrong PullHost format",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				pushFunc:  NewPushWithPortforwardFunc(nil, "", "", "", "<    >", nil),
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
			wantErr: clierror.Wrap(errors.New("registries must be valid RFC 3986 URI authorities: <    >"),
				clierror.New("failed to push image to the in-cluster registry")),
		},
		{
			name: "write image to in-cluster registry error",
			args: args{
				ctx:       context.Background(),
				imageName: "test:image",
				pushFunc: NewPushWithPortforwardFunc(nil, "podname", "podnamespace", "1234", "testhost:123", &basicAuth{
					username: "username",
					password: "password",
				}),
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
			wantErr: clierror.Wrap(errors.New("test error"),
				clierror.New("failed to push image to the in-cluster registry"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := importImage(tt.args.ctx, tt.args.imageName, tt.args.pushFunc, tt.args.utils)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}
