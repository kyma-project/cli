package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	dockerbuild "github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

var (
	testDockerfile        = `FROM alpine:latest`
	testWrongDockerignore = `$%^&!@()/\|[]{};:'',<.>=+-_1`
)

func TestBuild(t *testing.T) {

	t.Run("build image", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		require.NoError(t, os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm))

		mock := &dockerClientMock{}
		cli := NewTestClient(mock)

		err := cli.Build(context.Background(), BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})

		require.NoError(t, err)
	})

	t.Run("image build error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		require.NoError(t, os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm))

		mock := &dockerClientMock{err: errors.New("test error")}
		cli := NewTestClient(mock)

		err := cli.Build(context.Background(), BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})

		require.ErrorContains(t, err, "test error")
	})

	t.Run("wrong build response", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		require.NoError(t, os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm))

		mock := &dockerClientMock{
			bodyData: []byte{'}'},
		}
		cli := NewTestClient(mock)

		err := cli.Build(context.Background(), BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})

		require.ErrorContains(t, err, "invalid character '}'")
	})

	t.Run("wrong context error", func(t *testing.T) {
		tmpDir := t.TempDir()

		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		require.NoError(t, os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm))

		dockerignorePath := fmt.Sprintf("%s/.dockerignore", tmpDir)
		require.NoError(t, os.WriteFile(dockerignorePath, []byte(testWrongDockerignore), os.ModePerm))

		mock := &dockerClientMock{}
		cli := NewTestClient(mock)

		err := cli.Build(context.Background(), BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})

		require.ErrorContains(t, err, "error validating docker context")
	})

	t.Run("dockerfile not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)

		mock := &dockerClientMock{}
		cli := NewTestClient(mock)

		err := cli.Build(context.Background(), BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})

		require.ErrorContains(t, err, "no such file or directory")
	})
}

type dockerClientMock struct {
	client.Client
	err      error
	bodyData []byte
}

func (m *dockerClientMock) ImageBuild(ctx context.Context, buildContext io.Reader, options dockerbuild.ImageBuildOptions) (dockerbuild.ImageBuildResponse, error) {
	return dockerbuild.ImageBuildResponse{
		Body: io.NopCloser(bytes.NewReader(m.bodyData)),
	}, m.err
}
