package kustomize

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
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

func LifecycleManagerImageModifier(overrideString string, onOverride func(image string)) (imagetag.Filter, error) {
	override, err := parseOverride(overrideString)
	if err != nil {
		return imagetag.Filter{}, err
	}
	return ImageModifier(
		"*lifecycle-manager*", override.name, override.tag, override.digest,
		func(key, value, tag string, node *yaml.RNode) { onOverride(value) },
	), nil
}

type override struct {
	name   string
	digest string
	tag    string
}

var ErrImageInvalidArgs = errors.New(
	`invalid format of image, use one of the following options:
- <image>:<newtag>, in which case both image and tag get overritten
- <image>@<digest>, in which case both image and digest get overritten
- <tag>, in which case the default image is used but with a different tag`,
)

// parseOverride parses the override parameters
// from the given arg into a struct
// heavily inspired by kustomize edit https://github.com/kubernetes-sigs/kustomize/blob/22dbd3eb17d9980f900d761ff3665c2b1849726b/kustomize/commands/edit/set/setimage.go#L231
func parseOverride(arg string) (override, error) {
	// match <image>@<digest>
	if d := strings.Split(arg, "@"); len(d) > 1 {
		return override{
			name:   d[0],
			digest: d[1],
		}, nil
	}

	// match <image>:<tag>
	if t := regexp.MustCompile(`^(.*):([a-zA-Z0-9._-]*|\*)$`).FindStringSubmatch(arg); len(t) == 3 {
		return override{
			name: t[1],
			tag:  t[2],
		}, nil
	}

	// match <image>
	if len(arg) > 0 {
		return override{
			tag: arg,
		}, nil
	}
	return override{}, ErrImageInvalidArgs
}

func ControllerImageModifier(img, tag string) imagetag.Filter {
	return ImageModifier(ControllerImageName, img, tag, "", nil)
}

func ImageModifier(
	name, img, tag, digest string, callback func(key, value, tag string, node *yaml.RNode),
) imagetag.Filter {
	filter := imagetag.Filter{
		ImageTag: types.Image{
			Name:    name,
			NewName: img,
			NewTag:  tag,
			Digest:  digest,
		},
		FsSlice: []types.FieldSpec{
			{Path: "spec/containers[]/image"},
			{Path: "spec/template/spec/containers[]/image"},
		},
	}
	if callback != nil {
		(&filter).WithMutationTracker(callback)
	}
	return filter
}
