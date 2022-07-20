// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// *************************************************************************************
// Docker Manifest v2 Schema 1 Support
// see also:
// - https://docs.docker.com/registry/spec/manifest-v2-1/
// - https://github.com/moby/moby/blob/master/image/v1/imagev1.go
// - https://github.com/containerd/containerd/blob/main/remotes/docker/schema1/converter.go
// *************************************************************************************

const (
	MediaTypeDockerV2Schema1Manifest       = "application/vnd.docker.distribution.manifest.v1+json"
	MediaTypeDockerV2Schema1SignedManifest = images.MediaTypeDockerSchema1Manifest
	MediaTypeImageLayerZstd                = "application/vnd.oci.image.layer.v1.tar+zstd"
)

// FSLayer represents 1 item in a schema 1 "fsLayers" list
type FSLayer struct {
	BlobSum digest.Digest `json:"blobSum"`
}

// History represents 1 item in a schema 1 "history" list
type History struct {
	V1Compatibility string `json:"v1Compatibility"`
}

// V1Manifest describes a Docker v2 Schema 1 manifest
type V1Manifest struct {
	FSLayers []FSLayer `json:"fsLayers"`
	History  []History `json:"history"`
}

// V1History is the unmarshalled v1Compatibility property of a history item
type V1History struct {
	Author          string    `json:"author,omitempty"`
	Created         time.Time `json:"created"`
	Comment         string    `json:"comment,omitempty"`
	ThrowAway       *bool     `json:"throwaway,omitempty"`
	Size            *int      `json:"Size,omitempty"`
	ContainerConfig struct {
		Cmd []string `json:"Cmd,omitempty"`
	} `json:"container_config,omitempty"`
}

// ConvertV1ManifestToV2 converts a Docker v2 Schema 1 manifest to Docker v2 Schema 2.
// The converted manifest is stored in the cache. The descriptor of the cached manifest is returned.
func ConvertV1ManifestToV2(ctx context.Context, client Client, cache Cache, ref string, v1ManifestDesc ocispecv1.Descriptor) (ocispecv1.Descriptor, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := client.Fetch(ctx, ref, v1ManifestDesc, buf); err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to fetch v1 manifest blob: %w", err)
	}

	var v1Manifest V1Manifest
	if err := json.Unmarshal(buf.Bytes(), &v1Manifest); err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to unmarshal v1 manifest: %w", err)
	}

	layers, diffIDs, history, err := ParseV1Manifest(ctx, client, ref, &v1Manifest)
	if err != nil {
		return ocispecv1.Descriptor{}, err
	}

	v2ConfigDesc, v2ConfigBytes, err := CreateV2Config(&v1Manifest, diffIDs, history)
	if err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to create v2 config: %w", err)
	}

	v2ManifestDesc, v2ManifestBytes, err := CreateV2Manifest(v2ConfigDesc, layers)
	if err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to create v2 manifest: %w", err)
	}

	err = cache.Add(v2ConfigDesc, io.NopCloser(bytes.NewReader(v2ConfigBytes)))
	if err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to write config blob to cache: %w", err)
	}

	err = cache.Add(v2ManifestDesc, io.NopCloser(bytes.NewReader(v2ManifestBytes)))
	if err != nil {
		return ocispecv1.Descriptor{}, fmt.Errorf("unable to write manifest blob to cache: %w", err)
	}

	return v2ManifestDesc, nil
}

// ParseV1Manifest returns the data necessary to build a v2 manifest from a v1 manifest
func ParseV1Manifest(ctx context.Context, client Client, ref string, v1Manifest *V1Manifest) (layers []ocispecv1.Descriptor, diffIDs []digest.Digest, history []ocispecv1.History, err error) {
	layers = []ocispecv1.Descriptor{}
	diffIDs = []digest.Digest{}
	history = []ocispecv1.History{}

	// layers in v1 are reversed compared to v2 --> iterate backwards
	for i := len(v1Manifest.FSLayers) - 1; i >= 0; i-- {
		var h V1History
		if err := json.Unmarshal([]byte(v1Manifest.History[i].V1Compatibility), &h); err != nil {
			return nil, nil, nil, fmt.Errorf("unable to unmarshal v1 history: %w", err)
		}

		emptyLayer := isEmptyLayer(&h)

		hs := ocispecv1.History{
			Author:     h.Author,
			Comment:    h.Comment,
			Created:    &h.Created,
			CreatedBy:  strings.Join(h.ContainerConfig.Cmd, " "),
			EmptyLayer: emptyLayer,
		}
		history = append(history, hs)

		if !emptyLayer {
			fslayer := v1Manifest.FSLayers[i]
			layerDesc := ocispecv1.Descriptor{
				Digest: fslayer.BlobSum,
				Size:   -1,
			}

			buf := bytes.NewBuffer([]byte{})
			if err := client.Fetch(ctx, ref, layerDesc, buf); err != nil {
				return nil, nil, nil, fmt.Errorf("unable to fetch layer blob: %w", err)
			}
			data := buf.Bytes()

			decompressedReader, err := compression.DecompressStream(bytes.NewReader(data))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to decompress layer blob: %w", err)
			}

			decompressedData, err := ioutil.ReadAll(decompressedReader)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to read decompressed layer blob: %w", err)
			}

			var mediatype string
			switch decompressedReader.GetCompression() {
			case compression.Uncompressed:
				mediatype = ocispecv1.MediaTypeImageLayer
			case compression.Gzip:
				mediatype = ocispecv1.MediaTypeImageLayerGzip
			case compression.Zstd:
				mediatype = MediaTypeImageLayerZstd
			}

			des := ocispecv1.Descriptor{
				Digest:    fslayer.BlobSum,
				MediaType: mediatype,
				Size:      int64(len(data)),
			}

			layers = append(layers, des)
			diffIDs = append(diffIDs, digest.FromBytes(decompressedData))
		}
	}

	return
}

// CreateV2Manifest creates a v2 manifest
func CreateV2Manifest(configDesc ocispecv1.Descriptor, layers []ocispecv1.Descriptor) (ocispecv1.Descriptor, []byte, error) {
	v2Manifest := ocispecv1.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: configDesc,
		Layers: layers,
	}

	marshaledV2Manifest, err := json.Marshal(v2Manifest)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to marshal manifest: %w", err)
	}

	v2ManifestDesc := ocispecv1.Descriptor{
		MediaType: ocispecv1.MediaTypeImageManifest,
		Digest:    digest.FromBytes(marshaledV2Manifest),
		Size:      int64(len(marshaledV2Manifest)),
	}

	return v2ManifestDesc, marshaledV2Manifest, nil
}

// CreateV2Config creates a v2 config
func CreateV2Config(v1Manifest *V1Manifest, diffIDs []digest.Digest, history []ocispecv1.History) (ocispecv1.Descriptor, []byte, error) {
	var config map[string]*json.RawMessage
	if err := json.Unmarshal([]byte(v1Manifest.History[0].V1Compatibility), &config); err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to unmarshal config from v1 history: %w", err)
	}

	delete(config, "id")
	delete(config, "parent")
	delete(config, "Size")
	delete(config, "parent_id")
	delete(config, "layer_id")
	delete(config, "throwaway")

	rootfs := ocispecv1.RootFS{
		Type:    "layers",
		DiffIDs: diffIDs,
	}

	rootfsRaw, err := rawJSON(rootfs)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to convert rootfs to JSON: %w", err)
	}
	config["rootfs"] = rootfsRaw

	historyRaw, err := rawJSON(history)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to convert history to JSON: %w", err)
	}
	config["history"] = historyRaw

	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return ocispecv1.Descriptor{}, nil, fmt.Errorf("unable to marshal config: %w", err)
	}

	configDesc := ocispecv1.Descriptor{
		MediaType: ocispecv1.MediaTypeImageConfig,
		Digest:    digest.FromBytes(marshaledConfig),
		Size:      int64(len(marshaledConfig)),
	}

	return configDesc, marshaledConfig, nil
}

// isEmptyLayer returns whether the v1 compatibility history describes an empty layer.
// A return value of true indicates the layer is empty, however false does not indicate non-empty.
func isEmptyLayer(h *V1History) bool {
	// There doesn't seem to be a spec that describes whether the throwAway and size fields must exist or not.
	// At least in the Docker implementation, throwAway is optional (https://github.com/moby/moby/blob/master/distribution/pull_v2.go#L524).
	// For size we can only assume the same.
	// The whole logic could be interpreted as: "If clients which pushed the content made the effort to indicate
	// that a layer is empty, we can safely throw it away. For all other cases we copy every layer."
	if h.ThrowAway != nil {
		return *h.ThrowAway
	}
	if h.Size != nil {
		return *h.Size == 0
	}

	return false
}

// rawJSON converts an arbitrary value to json.RawMessage
func rawJSON(value interface{}) (*json.RawMessage, error) {
	jsonval, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&jsonval), nil
}
