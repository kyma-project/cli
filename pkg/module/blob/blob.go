package blob

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	pathutil "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/opencontainers/go-digest"
)

// MediaTypeTar defines the media type for a tarred file
const MediaTypeTar = "application/x-tar"

// MediaTypeGZip defines the media type for a gzipped file
const MediaTypeGZip = "application/gzip"

// MediaTypeOctetStream is the media type for any binary data.
const MediaTypeOctetStream = "application/octet-stream"

// Output is the output generated when reading a blob.Input.
type Output struct {
	mimeType string
	digest   string
	size     int64
	reader   io.ReadCloser
}

func (o *Output) Get() ([]byte, error) {
	return io.ReadAll(o.reader)
}

func (o *Output) Reader() (io.ReadCloser, error) {
	return o.reader, nil
}

func (o *Output) Close() error {
	return o.reader.Close()
}

func (o *Output) Size() int64 {
	return o.size
}

func (o *Output) MimeType() string {
	return o.mimeType
}

func (o *Output) DigestKnown() bool {
	return true
}

func (o *Output) Digest() digest.Digest {
	return digest.FromString(o.digest)
}

func (o *Output) Dup() (accessio.BlobAccess, error) {
	return o, nil
}

type InputType string

const (
	FileInputType = "file"
	DirInputType  = "dir"
)

// Input defines a local resource input that should be added to the component descriptor and
// to the resource's access.
type Input struct {
	// Type defines the input type of the blob to be added.
	// Note that an input blob of type directory is automatically tarred.
	Type InputType `json:"type"`
	// MediaType is the media type of the defined file that is also added to the OCI layer.
	// Should be a custom media type in the form of "application/vnd.<mydomain>.<my description>"
	MediaType string `json:"mediaType,omitempty"`
	// Path is the path that points to the blob to be added.
	Path string `json:"path"`
	// CompressWithGzip defines that the blob should be automatically compressed using gzip.
	CompressWithGzip *bool `json:"compress,omitempty"`
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only relevant for blob input type directory.
	PreserveDir bool `json:"preserveDir,omitempty"`
	// IncludeFiles is a list of shell file name patterns that describe the files that should be included.
	// If nothing is defined, all files are included.
	// Only relevant for blob input type directory.
	IncludeFiles []string `json:"includeFiles,omitempty"`
	// ExcludeFiles is a list of shell file name patterns that describe the files that should be excluded from the resulting tar.
	// Excluded files always overwrite included files.
	// Only relevant for blob input type directory.
	ExcludeFiles []string `json:"excludeFiles,omitempty"`
	// FollowSymlinks configures to follow and resolve symlinks when a directory is tarred.
	// This option will include the content of the symlink directly in the tar.
	// Use this option with care!
	FollowSymlinks bool `json:"followSymlinks,omitempty"`
}

// Compress returns wether the blob should be compressed using gzip.
func (input Input) Compress() bool {
	if input.CompressWithGzip == nil {
		return false
	}
	return *input.CompressWithGzip
}

// SetMediaTypeIfNotDefined sets the media type of the input blob if it's not defined.
func (input *Input) SetMediaTypeIfNotDefined(mediaType string) {
	if len(input.MediaType) != 0 {
		return
	}
	input.MediaType = mediaType
}

func AccessForFileOrFolder(fs vfs.FileSystem, input *Input) (accessio.BlobAccess, error) {
	return input.Read(context.Background(), fs)
}

// Read reads the configured blob and returns a reader to the given file.
func (input *Input) Read(ctx context.Context, fs vfs.FileSystem) (*Output, error) {
	inputInfo, err := fs.Stat(input.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to get info for input blob from %q, %w", input.Path, err)
	}

	// automatically tar the input artifact if it is a directory
	if input.Type == DirInputType {
		if !inputInfo.IsDir() {
			return nil, fmt.Errorf("resource type is dir, but you provided a file")
		}

		var data bytes.Buffer

		if input.Compress() {
			input.SetMediaTypeIfNotDefined(MediaTypeGZip)
			gw := gzip.NewWriter(&data)
			if err := TarFileSystem(
				ctx, fs, input.Path, gw, TarFileSystemOptions{
					IncludeFiles:   input.IncludeFiles,
					ExcludeFiles:   input.ExcludeFiles,
					PreserveDir:    input.PreserveDir,
					FollowSymlinks: input.FollowSymlinks,
				},
			); err != nil {
				return nil, fmt.Errorf("unable to tar input artifact: %w", err)
			}
			if err := gw.Close(); err != nil {
				return nil, fmt.Errorf("unable to close gzip writer: %w", err)
			}
		} else {
			input.SetMediaTypeIfNotDefined(MediaTypeTar)
			if err := TarFileSystem(
				ctx, fs, input.Path, &data, TarFileSystemOptions{
					IncludeFiles:   input.IncludeFiles,
					ExcludeFiles:   input.ExcludeFiles,
					PreserveDir:    input.PreserveDir,
					FollowSymlinks: input.FollowSymlinks,
				},
			); err != nil {
				return nil, fmt.Errorf("unable to tar input artifact: %w", err)
			}
		}

		return &Output{
			mimeType: input.MediaType,
			digest:   digest.FromBytes(data.Bytes()).String(),
			size:     int64(data.Len()),
			reader:   io.NopCloser(&data),
		}, nil
	} else if input.Type == FileInputType {
		if inputInfo.IsDir() {
			return nil, fmt.Errorf("resource type is file, but you provided a directory")
		}
		// otherwise just open the file
		inputBlob, err := fs.Open(input.Path)
		if err != nil {
			return nil, fmt.Errorf("unable to read input blob from %q: %w", input.Path, err)
		}
		blobDigest, err := digest.FromReader(inputBlob)
		if err != nil {
			return nil, fmt.Errorf("unable to calculate digest for input blob from %q, %w", input.Path, err)
		}
		if _, err := inputBlob.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("unable to reset input file: %s", err)
		}

		if input.Compress() {
			input.SetMediaTypeIfNotDefined(MediaTypeGZip)
			var data bytes.Buffer
			gw := gzip.NewWriter(&data)
			if _, err := io.Copy(gw, inputBlob); err != nil {
				return nil, fmt.Errorf("unable to compress input file %q: %w", input.Path, err)
			}
			if err := gw.Close(); err != nil {
				return nil, fmt.Errorf("unable to close gzip writer: %w", err)
			}

			return &Output{
				mimeType: input.MediaType,
				digest:   digest.FromBytes(data.Bytes()).String(),
				size:     int64(data.Len()),
				reader:   io.NopCloser(&data),
			}, nil
		}
		// default media type to binary data if nothing else is defined
		input.SetMediaTypeIfNotDefined(MediaTypeOctetStream)
		return &Output{
			mimeType: input.MediaType,
			digest:   blobDigest.String(),
			size:     inputInfo.Size(),
			reader:   inputBlob,
		}, nil
	} else {
		return nil, fmt.Errorf("unknown input type %q", input.Path)
	}
}

// TarFileSystemOptions describes additional options for tarring a filesystem.
type TarFileSystemOptions struct {
	IncludeFiles []string
	ExcludeFiles []string
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only supported for Type directory.
	PreserveDir    bool
	FollowSymlinks bool

	root string
}

// Included determines whether a file should be included.
func (opts *TarFileSystemOptions) Included(path string) (bool, error) {
	// Tf a root path is given, remove it from the path to be checked
	if len(opts.root) != 0 {
		path = strings.TrimPrefix(path, opts.root)
	}
	// First, check if an exclude regex matches
	for _, ex := range opts.ExcludeFiles {
		match, err := filepath.Match(ex, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", ex)
		}
		if match {
			return false, nil
		}
	}

	// if no includes are defined, include all files
	if len(opts.IncludeFiles) == 0 {
		return true, nil
	}
	// otherwise check if the file should be included
	for _, in := range opts.IncludeFiles {
		match, err := filepath.Match(in, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", in)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

// TarFileSystem creates a tar archive from a filesystem.
func TarFileSystem(
	ctx context.Context, fs vfs.FileSystem, root string, writer io.Writer, opts TarFileSystemOptions,
) error {
	tw := tar.NewWriter(writer)
	if opts.PreserveDir {
		opts.root = pathutil.Base(root)
	}
	if err := addFileToTar(ctx, fs, tw, opts.root, root, opts); err != nil {
		return err
	}
	return tw.Close()
}

func addFileToTar(
	ctx context.Context, fs vfs.FileSystem, tw *tar.Writer, path string, realPath string, opts TarFileSystemOptions,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	log := logr.FromContextOrDiscard(ctx)

	if len(path) != 0 { // do not check the root
		include, err := opts.Included(path)
		if err != nil {
			return err
		}
		if !include {
			return nil
		}
	}
	info, err := fs.Lstat(realPath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path

	// zero out time modifiers to always get the exact same hash on tar layers; only contents should influence the hash not when they were generated
	header.AccessTime = time.Time{}
	header.ChangeTime = time.Time{}
	header.ModTime = time.Time{}

	switch {
	case info.IsDir():
		// do not write root header
		if len(path) != 0 {
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write header for %q: %w", path, err)
			}
		}
		err := vfs.Walk(
			fs, realPath, func(subFilePath string, info os.FileInfo, err error) error {
				if subFilePath == realPath {
					return nil
				}
				if err != nil {
					return err
				}
				relPath, err := filepath.Rel(realPath, subFilePath)
				if err != nil {
					return fmt.Errorf("unable to calculate relative path for %s: %w", subFilePath, err)
				}
				return addFileToTar(ctx, fs, tw, pathutil.Join(path, relPath), subFilePath, opts)
			},
		)
		return err
	case info.Mode().IsRegular():
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %q: %w", path, err)
		}
		file, err := fs.OpenFile(realPath, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", path, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			copyErr := err
			err = file.Close()
			if err != nil {
				return fmt.Errorf("unable to close file %q: %w", path, err)
			}
			return fmt.Errorf("unable to add file to tar %q: %w", path, copyErr)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file %q: %w", path, err)
		}
		return nil
	case header.Typeflag == tar.TypeSymlink:
		if !opts.FollowSymlinks {
			log.Info(fmt.Sprintf("symlink found in %q but symlinks are not followed", path))
			return nil
		}
		realPath, err := vfs.EvalSymlinks(fs, realPath)
		if err != nil {
			return fmt.Errorf("unable to follow symlink %s: %w", realPath, err)
		}
		return addFileToTar(ctx, fs, tw, path, realPath, opts)
	default:
		return fmt.Errorf("unsupported file type %s in %s", info.Mode().String(), path)
	}
}
