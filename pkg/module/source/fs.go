package source

import (
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

const (
	sourceType      = "FileSystem"
	refFsNoGitLabel = "no-git.kyma-project.io/ref"
)

type FileSystemSource struct{}

func NewFileSystemSource() *FileSystemSource {
	return &FileSystemSource{}
}

func (f *FileSystemSource) FetchSource(_ cpi.Context, _, _, version string) (*ocm.Source, error) {
	label, err := ocmv1.NewLabel(refLabel, refFsNoGitLabel, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, err
	}

	return &ocm.Source{
		SourceMeta: ocm.SourceMeta{
			Type: sourceType,
			ElementMeta: ocm.ElementMeta{
				Name:    ocmIdentityName,
				Version: version,
				Labels:  ocmv1.Labels{*label},
			},
		},
	}, nil
}
