package scaffold

import (
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

const embeddedRoot = "resources"

//go:embed resources/*
var embeddedRes embed.FS

// builderOptions is used to keep all the top-level options that users may want to use for generating module files.
type builderOptions struct {
	ModuleName string
}

// dataTemplateOptions extends builderOptions with some per-file details, like a file's path
type dataTemplateOptions struct {
	builderOptions
	Path string
}

// resourceBuilder is used to create resources with data coming from files embedded along with the binary
type resourceBuilder struct {
	targetFs vfs.FileSystemWithWorkingDirectory
	opts     builderOptions
	err      error
}

func (rb *resourceBuilder) createDirectory(pathName string) *resourceBuilder {
	if rb.err != nil {
		return rb
	}

	pathName = strings.TrimPrefix(pathName, "/")

	err := rb.targetFs.Mkdir(pathName, directoryMode)
	if err != nil {
		rb.err = fmt.Errorf("An error while creating directory %q: %w", pathName, err)
	}

	return rb
}

// createFileFromTemplate creates a file in the target filesystem using a Golang template with data from the file with the same name in the embedded filesystem
// The template from the file is resolved against a dataTemplateOptions struct instance
func (rb *resourceBuilder) createFileFromTemplate(pathName string) *resourceBuilder {
	if rb.err != nil {
		return rb
	}

	pathName = strings.TrimPrefix(pathName, "/")
	embeddedPathName := path.Join(embeddedRoot, pathName)

	data, err := embeddedRes.ReadFile(embeddedPathName)
	if err != nil {
		rb.err = fmt.Errorf("Error while reading embedded file %q: %w", embeddedPathName, err)
		return rb
	}

	t, err := template.New("t").Parse(string(data))
	if err != nil {
		rb.err = fmt.Errorf("Error while parsing template from embedded file %q: %w", embeddedPathName, err)
		return rb
	}

	targetFile, err := rb.targetFs.Create(pathName)
	if err != nil {
		rb.err = fmt.Errorf("Error while creating target file %q: %w", pathName, err)
		return rb
	}
	defer targetFile.Close()

	templateOpts := dataTemplateOptions{
		builderOptions: rb.opts,
		Path:           pathName,
	}
	err = t.Execute(targetFile, templateOpts)
	if err != nil {
		rb.err = fmt.Errorf("Error while writing data to a target file %q: %w", pathName, err)
	}

	return rb
}

func (rb *resourceBuilder) result() error {
	return rb.err
}
