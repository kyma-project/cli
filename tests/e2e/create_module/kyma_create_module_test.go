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
	ocmOCIReg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/stretchr/testify/assert"
)

func Test_ModuleTemplate(t *testing.T) {
	ociRepoURL := os.Getenv("OCI_REPOSITORY_URL")
	testRepoURL := os.Getenv("TEST_REPOSITORY_URL")
	templatePath := os.Getenv("MODULE_TEMPLATE_PATH")

	template, err := e2e.ReadModuleTemplate(templatePath)
	assert.Nil(t, err)
	descriptor, err := template.GetDescriptor()
	assert.Nil(t, err)
	assert.Equal(t, descriptor.SchemaVersion(), v2.SchemaVersion)

	t.Run("test annotations", func(t *testing.T) {
		annotations := template.Annotations
		expectedModuleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
		assert.Equal(t, expectedModuleTemplateVersion, annotations["operator.kyma-project.io/module-version"])
		assert.Equal(t, "false", annotations["operator.kyma-project.io/is-cluster-scoped"])
	})

	t.Run("test descriptor.component.repositoryContexts", func(t *testing.T) {
		assert.Equal(t, 1, len(descriptor.RepositoryContexts))
		repo := descriptor.GetEffectiveRepositoryContext()
		assert.Equal(t, ociRepoURL, repo.Object["baseUrl"])
		assert.Equal(t, string(ocmOCIReg.OCIRegistryURLPathMapping), repo.Object["componentNameMapping"])
		assert.Equal(t, ocireg.Type, repo.Object["type"])
	})

	t.Run("test descriptor.component.resources[0]", func(t *testing.T) {
		assert.Equal(t, 2, len(descriptor.Resources))
		resource := descriptor.Resources[1]
		assert.Equal(t, module.RawManifestLayerName, resource.Name)
		assert.Equal(t, ocmMetaV1.LocalRelation, resource.Relation)
		assert.Equal(t, module.TypeYaml, resource.Type)
		expectedModuleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
		assert.Equal(t, expectedModuleTemplateVersion, resource.Version)
	})

	t.Run("test descriptor.component.resources[0].access", func(t *testing.T) {
		resourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
		assert.Nil(t, err)
		localBlobAccessSpec, ok := resourceAccessSpec.(*localblob.AccessSpec)
		assert.True(t, ok)
		assert.Equal(t, localblob.Type, localBlobAccessSpec.GetType())
		assert.Contains(t, "sha256:", localBlobAccessSpec.LocalReference)
	})

	t.Run("test descriptor.component.sources", func(t *testing.T) {
		assert.Equal(t, len(descriptor.Sources), 1)
		source := descriptor.Sources[0]
		sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
		assert.Nil(t, err)
		githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
		assert.True(t, ok)
		assert.Equal(t, github.Type, githubAccessSpec.Type)
		assert.Contains(t, githubAccessSpec.RepoURL, testRepoURL)
	})

	t.Run("test security scan labels", func(t *testing.T) {
		secScanLabels := descriptor.Sources[0].Labels
		assert.Equal(t, map[string]string{
			"git.kyma-project.io/ref":                  "refs/heads/main",
			"scan.security.kyma-project.io/dev-branch": "main",
			"scan.security.kyma-project.io/rc-tag":     "0.5.0",
			"scan.security.kyma-project.io/language":   "golang-mod",
			"scan.security.kyma-project.io/exclude":    "**/test/**,**/*_test.go",
		}, e2e.Flatten(secScanLabels))
	})
}
