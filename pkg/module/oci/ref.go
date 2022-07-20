package oci

import (
	"errors"
	"fmt"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
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
