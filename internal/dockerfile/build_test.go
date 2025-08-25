package dockerfile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	dockerbuild "github.com/docker/docker/api/types/build"
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
		err := os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm)
		require.NoError(t, err)

		builder := &imageBuilder{
			dockerClient: &dockerClientMock{},
			out:          io.Discard,
		}

		err = builder.do(context.Background(), &BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})
		require.NoError(t, err)
	})

	t.Run("image build error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		err := os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm)
		require.NoError(t, err)

		builder := &imageBuilder{
			dockerClient: &dockerClientMock{
				err: errors.New("test error"),
			},
			out: io.Discard,
		}

		err = builder.do(context.Background(), &BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})
		require.ErrorContains(t, err, "test error")
	})

	t.Run("wrong build response", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		err := os.WriteFile(dockerfilePath, []byte(testDockerfile), os.ModePerm)
		require.NoError(t, err)

		builder := &imageBuilder{
			dockerClient: &dockerClientMock{
				bodyData: []byte{'}'},
			},
			out: io.Discard,
		}

		err = builder.do(context.Background(), &BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})
		require.ErrorContains(t, err, "invalid character '}' looking for beginning of value")
	})

	t.Run("wrong context error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)
		dockerignorePath := fmt.Sprintf("%s/.dockerignore", tmpDir)
		err := os.WriteFile(dockerignorePath, []byte(testWrongDockerignore), os.ModePerm)
		require.NoError(t, err)

		builder := &imageBuilder{
			dockerClient: &dockerClientMock{},
			out:          io.Discard,
		}

		err = builder.do(context.Background(), &BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})
		require.ErrorContains(t, err, "error checking context: syntax error in pattern")
	})

	t.Run("dockerfile not found error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := fmt.Sprintf("%s/Dockerfile", tmpDir)

		builder := &imageBuilder{
			dockerClient: &dockerClientMock{
				bodyData: []byte{'}'},
			},
			out: io.Discard,
		}

		err := builder.do(context.Background(), &BuildOptions{
			ImageName:      "test-name",
			BuildContext:   tmpDir,
			DockerfilePath: dockerfilePath,
		})
		require.ErrorContains(t, err, "no such file or directory")
	})
}

type dockerClientMock struct {
	err      error
	bodyData []byte
}

func (m *dockerClientMock) ImageBuild(ctx context.Context, buildContext io.Reader, options dockerbuild.ImageBuildOptions) (dockerbuild.ImageBuildResponse, error) {
	return dockerbuild.ImageBuildResponse{
		Body: io.NopCloser(bytes.NewReader(m.bodyData)),
	}, m.err
}
