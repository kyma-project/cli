package create_module_test

import (
	"os"
	"testing"

	"github.com/kyma-project/cli/pkg/module"
	"github.com/kyma-project/cli/tests/e2e"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	ocmMetaV1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	v2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	ocmOCIReg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/stretchr/testify/assert"
)

func Test_ModuleTemplate(t *testing.T) {
	moduleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
	ociRepoURL := os.Getenv("OCI_REPOSITORY_URL")
	testRepoURL := os.Getenv("TEST_REPOSITORY_URL")

	template, err := e2e.ReadModuleTemplate(os.Getenv("MODULE_TEMPLATE_PATH"))
	assert.Nil(t, err)
	descriptor, err := template.GetDescriptor()
	assert.Nil(t, err)
	assert.Equal(t, descriptor.SchemaVersion(), v2.SchemaVersion)

	// test descriptor.component.repositoryContexts
	assert.Equal(t, len(descriptor.RepositoryContexts), 1)
	unstructuredRepo := descriptor.GetEffectiveRepositoryContext()
	typedRepo, err := unstructuredRepo.Evaluate(cpi.DefaultContext().RepositoryTypes())
	assert.Nil(t, err)
	concreteRepo, ok := typedRepo.(*genericocireg.RepositorySpec)
	assert.Equal(t, ok, true)
	assert.Equal(t, concreteRepo.ComponentNameMapping, ocmOCIReg.OCIRegistryURLPathMapping)
	assert.Equal(t, concreteRepo.GetType(), ocireg.Type)
	assert.Equal(t, concreteRepo.Name(), ociRepoURL)

	// test descriptor.component.resources[0]
	assert.Equal(t, len(descriptor.Resources), 1)
	resource := descriptor.Resources[0]
	assert.Equal(t, resource.Name, module.RawManifestLayerName)
	assert.Equal(t, resource.Relation, ocmMetaV1.LocalRelation)
	assert.Equal(t, resource.Type, module.TypeYaml)
	assert.Equal(t, resource.Version, moduleTemplateVersion)

	// test descriptor.component.resources[0].access
	resourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	assert.Nil(t, err)
	localblobAccessSpec, ok := resourceAccessSpec.(*localblob.AccessSpec)
	assert.Equal(t, ok, true)
	assert.Equal(t, localblobAccessSpec.GetType(), localblob.Type)
	assert.Contains(t, localblobAccessSpec.LocalReference, "sha256:")

	// test descriptor.component.sources
	assert.Equal(t, len(descriptor.Sources), 1)
	source := descriptor.Sources[0]
	sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
	assert.Nil(t, err)
	githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
	assert.Equal(t, ok, true)
	assert.Equal(t, githubAccessSpec.Type, github.Type)
	assert.Contains(t, testRepoURL, githubAccessSpec.RepoURL)

	// test security scan labels
	secScanLabels := descriptor.Sources[0].Labels
	assert.Equal(t, map[string]string{
		"git.kyma-project.io/ref":                  "refs/heads/sec_scanners_change",
		"scan.security.kyma-project.io/dev-branch": "main",
		"scan.security.kyma-project.io/rc-tag":     "0.5.0",
		"scan.security.kyma-project.io/language":   "golang-mod",
		"scan.security.kyma-project.io/exclude":    "**/test/**,**/*_test.go",
	}, e2e.FlattenSecurityLabels(secScanLabels))
}
