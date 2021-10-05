package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	dockerConfigFile "github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	imageTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/kyma-project/cli/pkg/docker/mocks"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func Test_SplitDockerDomain(t *testing.T) {
	t.Parallel()
	test1 := "localhost:5000/test/testImage:1"
	d1, r1 := splitDockerDomain(test1)
	require.Equal(t, d1, "localhost:5000")
	require.Equal(t, r1, "test/testImage:1")

	test2 := "eu.gcr.io/test/testImage"
	d2, r2 := splitDockerDomain(test2)
	require.Equal(t, d2, "eu.gcr.io")
	require.Equal(t, r2, "test/testImage")

	test3 := "testImage"
	d3, r3 := splitDockerDomain(test3)
	require.Equal(t, d3, "index.docker.io")
	require.Equal(t, r3, "testImage")
}

func genConfigFile() *dockerConfigFile.ConfigFile {
	configFile := dockerConfigFile.New("tmpConfig")

	exampleAuth := types.AuthConfig{
		Username:      "user",
		Password:      "pass",
		Auth:          "",
		ServerAddress: "1.2.3.4",
		Email:         "foo@bar.com",
		IdentityToken: "identityFoo",
		RegistryToken: "registryFoo",
	}

	authStr := exampleAuth.Username + ":" + exampleAuth.Password

	msg := []byte(authStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	exampleAuth.Auth = string(encoded)
	configFile.AuthConfigs["example.com"] = exampleAuth
	return configFile
}

func Test_Resolve_happy_path(t *testing.T) {
	tmpHome, err := ioutil.TempDir("/tmp", "config-test")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpHome)

	configFile := genConfigFile()
	b, err := json.Marshal(configFile)
	assert.NilError(t, err)
	tmpFile := fmt.Sprintf("%s/config.json", tmpHome)
	err = ioutil.WriteFile(tmpFile, b, 0600)
	assert.NilError(t, err)

	os.Setenv("DOCKER_CONFIG", tmpHome)

	dockerCFG, err := resolve("example.com")
	assert.NilError(t, err)
	assert.Equal(t, dockerCFG.Username, "user")
	assert.Equal(t, dockerCFG.Password, "pass")
}

func Test_BuildKymaInstaller(t *testing.T) {
	t.Parallel()
	imageName := "kyma-project-foo"
	fooLocalSrcPath := "foo"

	// mocks
	mockDocker := &mocks.Client{}
	// mockKymaDocker := mocks.KymaDockerService{}
	stringReader := strings.NewReader("foo")
	fooReadCloser := ioutil.NopCloser(stringReader)

	fooArchiveTarOptions := &archive.TarOptions{}

	k := kymaDockerClient{
		Docker: mockDocker,
	}

	mockDocker.On("ArchiveDirectory", fooLocalSrcPath, fooArchiveTarOptions).Return(fooReadCloser, nil)
	// as context.deadline can have different clocks assume mock.anything here
	mockDocker.On("NegotiateAPIVersion", mock.Anything).Return(nil)
	fooArgs := make(map[string]*string)
	fooImageBuildOptions := imageTypes.ImageBuildOptions{
		Tags:           []string{strings.TrimSpace(string(imageName))},
		SuppressOutput: true,
		Remove:         true,
		Dockerfile:     path.Join("tools", "kyma-installer", "kyma.Dockerfile"),
		BuildArgs:      fooArgs,
	}
	fooImageBuildRes := imageTypes.ImageBuildResponse{
		Body:   fooReadCloser,
		OSType: "fooUnix",
	}
	mockDocker.On("ImageBuild", mock.Anything, fooReadCloser, fooImageBuildOptions).Return(fooImageBuildRes, nil)

	// test the function
	err := k.BuildKymaInstaller(fooLocalSrcPath, imageName)
	assert.NilError(t, err)
}

func Test_PushKymaInstaller(t *testing.T) {
	tmpHome, err := ioutil.TempDir("/tmp", "config-pus-kyma-test")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpHome)

	configFile := genConfigFile()
	b, err := json.Marshal(configFile)
	assert.NilError(t, err)
	tmpFile := fmt.Sprintf("%s/config.json", tmpHome)
	err = ioutil.WriteFile(tmpFile, b, 0600)
	assert.NilError(t, err)

	os.Setenv("DOCKER_CONFIG", tmpHome)
	image := "example.com/foo"

	// mocks
	mockDocker := &mocks.Client{}
	// as context.deadline can have different clocks assume mock.anything here

	k := kymaDockerClient{
		Docker: mockDocker,
	}
	mockDocker.On("NegotiateAPIVersion", mock.Anything).Return(nil)

	expectedAuth := types.AuthConfig{
		Username:      "user",
		Password:      "pass",
		IdentityToken: "identityFoo",
		RegistryToken: "registryFoo",
	}
	encodedJSON, _ := json.Marshal(expectedAuth)
	fooAuthStr := base64.URLEncoding.EncodeToString(encodedJSON)
	imagePushOptions := imageTypes.ImagePushOptions{RegistryAuth: fooAuthStr}
	stringReader := strings.NewReader("foo")
	fooReadCloser := ioutil.NopCloser(stringReader)

	var step step.Factory
	currentStep := step.NewStep("push kyma installer test")

	mockDocker.On("ImagePush", mock.Anything, image, imagePushOptions).Return(fooReadCloser, nil)

	err = k.PushKymaInstaller(image, currentStep)
	assert.NilError(t, err)

}

func Test_portSet(t *testing.T) {
	actual := portSet(map[string]string{"3000": "3001"})
	expected := nat.PortSet{"3000": struct{}{}}
	require.Equal(t, expected, actual)
}

func Test_mapMap(t *testing.T) {
	actual := portMap(map[string]string{"3000": "3001"})
	expected := nat.PortMap{"3000": []nat.PortBinding{
		{
			HostPort: "3001",
		},
	}}
	require.Equal(t, expected, actual)
}

func Test_PullImageAndStartContainergst(t *testing.T) {
	mockDocker := &mocks.Client{}
	mockWrapper := dockerWrapper{Docker: mockDocker}

	testOpts := ContainerRunOpts{
		ContainerName: "container-name",
		Envs:          nil,
		Image:         "valid-image",
		Ports:         map[string]string{"3000": "3001"},
	}
	testContainerID := "container-id-123"
	testErr := errors.New("container create and start error")
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		mockDocker.On("ImagePull", ctx, testOpts.Image, mock.AnythingOfType("types.ImagePullOptions")).Return(
			ioutil.NopCloser(bytes.NewReader(nil)), nil).Times(1)

		mockDocker.On("ContainerCreate", ctx, mock.AnythingOfType("*container.Config"),
			mock.AnythingOfType("*container.HostConfig"), mock.Anything, mock.Anything, testOpts.ContainerName).Return(
			container.ContainerCreateCreatedBody{ID: testContainerID}, nil).Times(1)

		mockDocker.On("ContainerStart", ctx, testContainerID, mock.AnythingOfType("types.ContainerStartOptions")).Return(nil).Times(1)

		id, err := mockWrapper.PullImageAndStartContainer(ctx, testOpts)

		require.Nil(t, err)
		require.Equal(t, testContainerID, id)
	})

	t.Run("image pull error", func(t *testing.T) {
		mockDocker.On("ImagePull", ctx, testOpts.Image, mock.AnythingOfType("types.ImagePullOptions")).Return(
			ioutil.NopCloser(bytes.NewReader(nil)), testErr).Times(1)

		id, err := mockWrapper.PullImageAndStartContainer(ctx, testOpts)

		require.Equal(t, "", id)
		require.NotNil(t, err)
		require.Equal(t, err, testErr)
	})

	t.Run("container create error", func(t *testing.T) {
		mockDocker.On("ImagePull", ctx, testOpts.Image, mock.AnythingOfType("types.ImagePullOptions")).Return(
			ioutil.NopCloser(bytes.NewReader(nil)), nil).Times(1)

		mockDocker.On("ContainerCreate", ctx, mock.AnythingOfType("*container.Config"),
			mock.AnythingOfType("*container.HostConfig"), mock.Anything, mock.Anything, testOpts.ContainerName).Return(
			container.ContainerCreateCreatedBody{}, testErr).Times(1)

		id, err := mockWrapper.PullImageAndStartContainer(ctx, testOpts)

		require.Equal(t, "", id)
		require.NotNil(t, err)
		require.Equal(t, err, testErr)
	})

	t.Run("container start error", func(t *testing.T) {
		mockDocker.On("ImagePull", ctx, testOpts.Image, mock.AnythingOfType("types.ImagePullOptions")).Return(
			ioutil.NopCloser(bytes.NewReader(nil)), nil).Times(1)

		mockDocker.On("ContainerCreate", ctx, mock.AnythingOfType("*container.Config"),
			mock.AnythingOfType("*container.HostConfig"), mock.Anything, mock.Anything, testOpts.ContainerName).Return(
			container.ContainerCreateCreatedBody{ID: testContainerID}, nil).Times(1)

		mockDocker.On("ContainerStart", ctx, testContainerID, mock.AnythingOfType("types.ContainerStartOptions")).Return(testErr).Times(1)

		id, err := mockWrapper.PullImageAndStartContainer(ctx, testOpts)

		require.Equal(t, "", id)
		require.NotNil(t, err)
		require.Equal(t, err, testErr)
	})
}

func Test_IsDockerDesktopOS(t *testing.T) {
	mockDocker := &mocks.Client{}
	mockWrapper := dockerWrapper{Docker: mockDocker}

	ctx := context.Background()
	dockerDesktopOSInfo := "Docker Desktop"
	testErr := errors.New("docker info error")

	t.Run("is docker desktop os", func(t *testing.T) {
		mockDocker.On("Info", ctx).Return(imageTypes.Info{OperatingSystem: dockerDesktopOSInfo}, nil).Times(1)

		dockerDesktop, err := mockWrapper.IsDockerDesktopOS(ctx)

		require.Nil(t, err)
		require.True(t, dockerDesktop)
	})

	t.Run("is not docker desktop os", func(t *testing.T) {
		mockDocker.On("Info", ctx).Return(imageTypes.Info{OperatingSystem: "not Docker Desktop"}, nil).Times(1)

		dockerDesktop, err := mockWrapper.IsDockerDesktopOS(ctx)

		require.Nil(t, err)
		require.False(t, dockerDesktop)
	})

	t.Run("info error", func(t *testing.T) {
		mockDocker.On("Info", ctx).Return(imageTypes.Info{OperatingSystem: "not Docker Desktop"}, testErr).Times(1)

		dockerDesktop, err := mockWrapper.IsDockerDesktopOS(ctx)

		require.NotNil(t, err)
		require.Equal(t, testErr, err)
		require.False(t, dockerDesktop)
	})
}
