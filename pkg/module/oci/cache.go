package oci

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// ErrNotFound is an error that indicates that the file is not cached
	ErrNotFound = errors.New("not cached")
)

// Cache is the interface for an OCI cache where descriptors can be added and fetched
type Cache interface {
	io.Closer
	Store
	Add(desc ocispecv1.Descriptor, reader io.ReadCloser) error
}

// Store describes a read-only descriptor store
type Store interface {
	Get(desc ocispecv1.Descriptor) (io.ReadCloser, error)
}

type inmemoryCache struct {
	store map[string][]byte
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache() Cache {
	return &inmemoryCache{
		store: make(map[string][]byte),
	}
}

func (fs *inmemoryCache) Close() error {
	return nil
}

func (fs *inmemoryCache) Get(desc ocispecv1.Descriptor) (io.ReadCloser, error) {
	data, ok := fs.store[desc.Digest.String()]
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(bytes.NewBuffer(data)), nil
}

func (fs *inmemoryCache) Add(desc ocispecv1.Descriptor, reader io.ReadCloser) error {
	if _, ok := fs.store[desc.Digest.String()]; ok {
		// already cached
		return nil
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return fmt.Errorf("unable to read data: %w", err)
	}
	fs.store[desc.Digest.String()] = buf.Bytes()
	return nil
}
