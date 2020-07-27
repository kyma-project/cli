package installation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	dockerConfigFile "github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func Test_SplitDockerDomain(t *testing.T) {
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

func Test_Resolve_happy_path(t *testing.T) {

	tmpHome, err := ioutil.TempDir("/tmp", "config-test")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpHome)

	configFile := dockerConfigFile.New("tmpConfig")
	exampleAuth := types.AuthConfig{
		Username:      "user",
		Password:      "pass",
		Auth:          "fooAuth",
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

	configFile.AuthConfigs["example.com/foo"] = exampleAuth
	b, err := json.Marshal(configFile)
	assert.NilError(t, err)
	tmpFile := fmt.Sprintf("%s/config.json", tmpHome)
	ioutil.WriteFile(tmpFile, b, 0644)

	os.Setenv("DOCKER_CONFIG", tmpHome)

	dockerCFG, err := resolve("example.com/foo")
	assert.NilError(t, err)
	assert.Equal(t, dockerCFG.Username, exampleAuth.Username)
	assert.Equal(t, dockerCFG.Password, exampleAuth.Password)
}

func Test_Resolve_no_file(t *testing.T) {
	os.Setenv("DOCKER_CONFIG", "file-not-exist")
	_, err := resolve("example.com/foo")
	assert.ErrorContains(t, err, "file not found")
}

func Test_BuildKymaInstaller(t *testing.T) {
	i := Installation{
		Options: &Options{
			IsLocal: false,
		},
	}
	err := i.buildKymaInstaller("foo")
	assert.NilError(t, err)
}
