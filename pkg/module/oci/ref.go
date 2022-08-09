package oci

import (
	"errors"
	"fmt"
	"path"
	"strings"

	dockerreference "github.com/containerd/containerd/reference/docker"
	"github.com/opencontainers/go-digest"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
)

// to find a suitable secret for images on Docker Hub, we need its two domains to do matching
const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"
)

// Ref creates an absolute OCI URL to the component descriptor with name and version at the registry in the given repository.
func Ref(repository cdv2.Repository, name, version string) (string, error) {
	repoCtx, err := GetOCIRepositoryContext(repository)
	if err != nil {
		return "", err
	}
	return cdoci.OCIRef(repoCtx, name, version)
}

// GetOCIRepositoryContext returns a OCIRegistryRepository from a repository
func GetOCIRepositoryContext(repoCtx cdv2.Repository) (cdv2.OCIRegistryRepository, error) {
	if repoCtx == nil {
		return cdv2.OCIRegistryRepository{}, errors.New("no repository provided")
	}
	var repo cdv2.OCIRegistryRepository
	switch r := repoCtx.(type) {
	case *cdv2.UnstructuredTypedObject:
		if err := r.DecodeInto(&repo); err != nil {
			return cdv2.OCIRegistryRepository{}, err
		}
	case *cdv2.OCIRegistryRepository:
		repo = *r
	default:
		return cdv2.OCIRegistryRepository{}, fmt.Errorf("unknown repository context type %s", repoCtx.GetType())
	}
	return repo, nil
}

// ParseRef parses a oci reference into a internal representation.
func ParseRef(ref string) (RefSpec, error) {
	if strings.Contains(ref, "://") {
		// remove protocol if exists
		i := strings.Index(ref, "://") + 3
		ref = ref[i:]
	}

	parsedRef, err := dockerreference.ParseDockerRef(ref)
	if err != nil {
		return RefSpec{}, err
	}

	spec := RefSpec{
		Host:       dockerreference.Domain(parsedRef),
		Repository: dockerreference.Path(parsedRef),
	}

	switch r := parsedRef.(type) {
	case dockerreference.Tagged:
		tag := r.Tag()
		spec.Tag = &tag
	case dockerreference.Digested:
		d := r.Digest()
		spec.Digest = &d
	}

	// fallback to legacy docker domain if applicable
	// this is how containerd translates the old domain for DockerHub to the new one, taken from containerd/reference/docker/reference.go:674
	if spec.Host == dockerHubDomain {
		spec.Host = dockerHubLegacyDomain
	}
	return spec, nil
}

// RefSpec is a go internal representation of a oci reference.
type RefSpec struct {
	// Host is the hostname of a oci ref.
	Host string
	// Repository is the part of a reference without its hostname
	Repository string
	// +optional
	Tag *string
	// +optional
	Digest *digest.Digest
}

func (r RefSpec) String() string {
	if r.Tag != nil {
		return fmt.Sprintf("%s:%s", r.Name(), *r.Tag)
	}
	if r.Digest != nil {
		return fmt.Sprintf("%s@%s", r.Name(), r.Digest.String())
	}
	return ""
}

func (r *RefSpec) Name() string {
	return path.Join(r.Host, r.Repository)
}

func (r *RefSpec) ShortName() string {
	t := strings.Split(r.Repository, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}
