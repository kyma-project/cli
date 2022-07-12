package blob

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	pathutil "path"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

// MediaTypeTar defines the media type for a tarred file
const MediaTypeTar = "application/x-tar"

// MediaTypeGZip defines the media type for a gzipped file
const MediaTypeGZip = "application/gzip"

// MediaTypeOctetStream is the media type for any binary data.
const MediaTypeOctetStream = "application/octet-stream"

// BlobOutput is the output if read BlobInput.
type BlobOutput struct {
	Digest string
	Size   int64
	Reader io.ReadCloser
}

type BlobInputType string

const (
	FileInputType = "file"
	DirInputType  = "dir"
)

// BlobInput defines a local resource input that should be added to the component descriptor and
// to the resource's access.
type BlobInput struct {
	// Type defines the input type of the blob to be added.
	// Note that a input blob of type "dir" is automatically tarred.
	Type BlobInputType `json:"type"`
	// MediaType is the mediatype of the defined file that is also added to the oci layer.
	// Should be a custom media type in the form of "application/vnd.<mydomain>.<my description>"
	MediaType string `json:"mediaType,omitempty"`
	// Path is the path that points to the blob to be added.
	Path string `json:"path"`
	// CompressWithGzip defines that the blob should be automatically compressed using gzip.
	CompressWithGzip *bool `json:"compress,omitempty"`
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only supported for Type dir.
	PreserveDir bool `json:"preserveDir,omitempty"`
	// IncludeFiles is a list of shell file name patterns that describe the files that should be included.
	// If nothing is defined all files are included.
	// Only relevant for blobinput type "dir".
	IncludeFiles []string `json:"includeFiles,omitempty"`
	// ExcludeFiles is a list of shell file name patterns that describe the files that should be excluded from the resulting tar.
	// Excluded files always overwrite included files.
	// Only relevant for blobinput type "dir".
	ExcludeFiles []string `json:"excludeFiles,omitempty"`
	// FollowSymlinks configures to follow and resolve symlinks when a directory is tarred.
	// This options will include the content of the symlink directly in the tar.
	// This option should be used with care.
	FollowSymlinks bool `json:"followSymlinks,omitempty"`
}

// Compress returns if the blob should be compressed using gzip.
func (input BlobInput) Compress() bool {
	if input.CompressWithGzip == nil {
		return false
	}
	return *input.CompressWithGzip
}

// SetMediaTypeIfNotDefined sets the media type of the input blob if its not defined
func (input *BlobInput) SetMediaTypeIfNotDefined(mediaType string) {
	if len(input.MediaType) != 0 {
		return
	}
	input.MediaType = mediaType
}

// Read reads the configured blob and returns a reader to the given file.
func (input *BlobInput) Read(ctx context.Context, fs vfs.FileSystem) (*BlobOutput, error) {
	inputInfo, err := fs.Stat(input.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to get info for input blob from %q, %w", input.Path, err)
	}

	// automatically tar the input artifact if it is a directory
	if input.Type == DirInputType {
		if !inputInfo.IsDir() {
			return nil, fmt.Errorf("resource type is dir but a file was provided")
		}

		var data bytes.Buffer

		if input.Compress() {
			input.SetMediaTypeIfNotDefined(MediaTypeGZip)
			gw := gzip.NewWriter(&data)
			if err := TarFileSystem(ctx, fs, input.Path, gw, TarFileSystemOptions{
				IncludeFiles:   input.IncludeFiles,
				ExcludeFiles:   input.ExcludeFiles,
				PreserveDir:    input.PreserveDir,
				FollowSymlinks: input.FollowSymlinks,
			}); err != nil {
				return nil, fmt.Errorf("unable to tar input artifact: %w", err)
			}
			if err := gw.Close(); err != nil {
				return nil, fmt.Errorf("unable to close gzip writer: %w", err)
			}
		} else {
			input.SetMediaTypeIfNotDefined(MediaTypeTar)
			if err := TarFileSystem(ctx, fs, input.Path, &data, TarFileSystemOptions{
				IncludeFiles:   input.IncludeFiles,
				ExcludeFiles:   input.ExcludeFiles,
				PreserveDir:    input.PreserveDir,
				FollowSymlinks: input.FollowSymlinks,
			}); err != nil {
				return nil, fmt.Errorf("unable to tar input artifact: %w", err)
			}
		}

		return &BlobOutput{
			Digest: digest.FromBytes(data.Bytes()).String(),
			Size:   int64(data.Len()),
			Reader: ioutil.NopCloser(&data),
		}, nil
	} else if input.Type == FileInputType {
		if inputInfo.IsDir() {
			return nil, fmt.Errorf("resource type is file but a directory was provided")
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

			return &BlobOutput{
				Digest: digest.FromBytes(data.Bytes()).String(),
				Size:   int64(data.Len()),
				Reader: ioutil.NopCloser(&data),
			}, nil
		}
		return &BlobOutput{
			Digest: blobDigest.String(),
			Size:   inputInfo.Size(),
			Reader: inputBlob,
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
	// Only supported for Type dir.
	PreserveDir    bool
	FollowSymlinks bool

	root string
}

// Included determines whether a file should be included.
func (opts *TarFileSystemOptions) Included(path string) (bool, error) {
	// if a root path is given remove it rom the path to be checked
	if len(opts.root) != 0 {
		path = strings.TrimPrefix(path, opts.root)
	}
	// first check if a exclude regex matches
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
func TarFileSystem(ctx context.Context, fs vfs.FileSystem, root string, writer io.Writer, opts TarFileSystemOptions) error {
	tw := tar.NewWriter(writer)
	if opts.PreserveDir {
		opts.root = pathutil.Base(root)
	}
	if err := addFileToTar(ctx, fs, tw, opts.root, root, opts); err != nil {
		return err
	}
	return tw.Close()
}

func addFileToTar(ctx context.Context, fs vfs.FileSystem, tw *tar.Writer, path string, realPath string, opts TarFileSystemOptions) error {
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

	switch {
	case info.IsDir():
		// do not write root header
		if len(path) != 0 {
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write header for %q: %w", path, err)
			}
		}
		err := vfs.Walk(fs, realPath, func(subFilePath string, info os.FileInfo, err error) error {
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
		})
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
			_ = file.Close()
			return fmt.Errorf("unable to add file to tar %q: %w", path, err)
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
