package kustomize

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	buildURLPattern = "%s?ref=%s" // pattern for URL locations Definition.Location?ref=Definition.Ref
	defaultURLRef   = "main"
	localRef        = "local"
)

type Definition struct {
	Name     string
	Ref      string
	Location string
}

func ParseKustomization(s string) (Definition, error) {
	// split URL from ref
	items := strings.Split(s, "@")
	if len(items) == 0 || len(items) > 2 {
		return Definition{}, fmt.Errorf(
			"the given kustomization %q could not be parsed: at least, it must contain a location (URL or path); optionally, URLs can have a reference in format URL@ref",
			s,
		)
	}

	res := Definition{}
	u, err := url.Parse(items[0])
	if err != nil {
		return Definition{}, fmt.Errorf(
			"could not parse the given location %q: make sure it is a valid URL or path", items[0],
		)
	}

	// URL case
	if u.Scheme != "" && u.Host != "" {
		pathChunks := strings.Split(u.Path, "/")
		if len(pathChunks) < 3 {
			return Definition{}, fmt.Errorf(
				"The provided URL %q does not belong to a repository. It must follow the format DOMAIN.EXT/OWNER/REPO/[SUBPATH]",
				items[0],
			)
		}
		res.Name = pathChunks[2]
		if len(items) == 2 {
			res.Ref = items[1]
		} else {
			res.Ref = defaultURLRef
		}
		res.Location = items[0]
	} else { // Path case
		res.Name = items[0]
		res.Ref = localRef
		res.Location = items[0]
	}

	return res, nil
}

const NoOutputFile = ""

// Build generates a manifest given a path using kustomize
// Additional args might be given to the kustomize build command
func Build(def Definition, outPath string, filters ...kio.Filter) ([]byte, error) {
	opts := krusty.MakeDefaultOptions()
	kustomize := krusty.MakeKustomizer(opts)

	path := def.Location
	if def.Ref != localRef {
		path = fmt.Sprintf(buildURLPattern, def.Location, def.Ref)
	}

	results, err := kustomize.Run(filesys.MakeFsOnDisk(), path)
	if err != nil {
		return nil, fmt.Errorf("could not build kustomization: %w", err)
	}

	for i, filter := range filters {
		if err := results.ApplyFilter(filter); err != nil {
			return nil, fmt.Errorf("could not apply filter (number %v): %w", i, err)
		}
	}

	yaml, err := results.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("could not parse kustomization as yaml: %w", err)
	}

	if outPath != "" {
		if err := os.WriteFile(outPath, yaml, os.ModePerm); err != nil {
			return nil, fmt.Errorf("could not write rendered kustomization as yaml to %s: %w", outPath, err)
		}
	}

	return yaml, nil
}

const (
	ControllerImageName = "controller"
)

func ControllerImageModifier(img, ver string) imagetag.Filter {
	return ImageModifier(ControllerImageName, img, ver, false, nil)
}

func ImageModifier(
	name, img, ver string, isDigest bool, callback func(key, value, tag string, node *yaml.RNode),
) imagetag.Filter {
	filter := imagetag.Filter{
		ImageTag: types.Image{
			Name:    name,
			NewName: img,
		},
		FsSlice: []types.FieldSpec{
			{Path: "spec/containers[]/image"},
			{Path: "spec/template/spec/containers[]/image"},
		},
	}
	if isDigest {
		filter.ImageTag.Digest = ver
	} else {
		filter.ImageTag.NewTag = ver
	}
	if callback != nil {
		(&filter).WithMutationTracker(callback)
	}
	return filter
}
