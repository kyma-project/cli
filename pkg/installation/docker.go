package installation

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	dockerConfig "github.com/docker/cli/cli/config"
	configTypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/kyma-project/cli/internal/minikube"
)

const (
	defaultRegistry = "index.docker.io"
)

// DockerErrorMessage is used to parse error messages coming from Docker
type DockerErrorMessage struct {
	Error string
}

func (i *Installation) buildKymaInstaller(imageName string) error {
	var dc *docker.Client
	var err error
	if i.Options.IsLocal {
		dc, err = minikube.DockerClient(i.Options.Verbose, i.Options.LocalCluster.Profile, i.Options.Timeout)
	} else {
		dc, err = docker.NewClientWithOpts(docker.FromEnv)
	}
	if err != nil {
		return err
	}

	reader, err := archive.TarWithOptions(i.Options.LocalSrcPath, &archive.TarOptions{})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()

	dc.NegotiateAPIVersion(ctx)

	args := make(map[string]*string)
	_, err = dc.ImageBuild(
		ctx,
		reader,
		types.ImageBuildOptions{
			Tags:           []string{strings.TrimSpace(string(imageName))},
			SuppressOutput: true,
			Remove:         true,
			Dockerfile:     path.Join("tools", "kyma-installer", "kyma.Dockerfile"),
			BuildArgs:      args,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *Installation) pushKymaInstaller() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()

	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return err
	}

	dc.NegotiateAPIVersion(ctx)

	domain, _ := splitDockerDomain(i.Options.CustomImage)
	auth, err := resolve(domain)
	if err != nil {
		return err
	}

	encodedJSON, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	pusher, err := dc.ImagePush(ctx, i.Options.CustomImage, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}

	defer pusher.Close()

	var errorMessage DockerErrorMessage
	buffIOReader := bufio.NewReader(pusher)

	for {
		streamBytes, err := buffIOReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		err = json.Unmarshal(streamBytes, &errorMessage)
		if err != nil {
			return err
		}
		if errorMessage.Error != "" {
			if strings.Contains(errorMessage.Error, "unauthorized") {
				return fmt.Errorf("failed to push Docker image: %s\nPlease run `docker login`", errorMessage.Error)
			}
			return fmt.Errorf("failed to push Docker image: %s", errorMessage.Error)
		}
	}

	return nil
}

func splitDockerDomain(name string) (domain, remainder string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = defaultRegistry, name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	return
}

// resolve finds the Docker credentials to push an image.
func resolve(registry string) (*configTypes.AuthConfig, error) {
	cf, err := dockerConfig.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return nil, err
	}

	if registry == defaultRegistry {
		registry = "https://" + defaultRegistry + "/v1/"
	}

	cfg, err := cf.GetAuthConfig(registry)
	if err != nil {
		return nil, err
	}

	empty := configTypes.AuthConfig{}
	if cfg == empty {
		return &empty, nil
	}
	return &configTypes.AuthConfig{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}, nil
}
