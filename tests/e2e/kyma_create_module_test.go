package e2e_test

import (
	"os"

	"github.com/kyma-project/cli/pkg/module"
	"github.com/kyma-project/cli/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	ocmMetaV1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	v2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	ocmOCIReg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Module Creation", Ordered, func() {
	moduleTemplateVersion := os.Getenv("MODULE_TEMPLATE_VERSION")
	ociRepoURL := os.Getenv("OCI_REPOSITORY_URL")
	testRepoURL := os.Getenv("TEST_REPOSITORY_URL")

	template, err := e2e.ReadModuleTemplate(os.Getenv("MODULE_TEMPLATE_PATH"))
	Expect(err).To(Not(HaveOccurred()))
	descriptor, err := template.GetDescriptor()
	Expect(err).To(Not(HaveOccurred()))
	Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))

	It("Then descriptor.component.repositoryContexts should be correct", func() {
		Expect(len(descriptor.RepositoryContexts)).To(Equal(1))
		unstructuredRepo := descriptor.GetEffectiveRepositoryContext()
		typedRepo, err := unstructuredRepo.Evaluate(cpi.DefaultContext().RepositoryTypes())
		Expect(err).To(Not(HaveOccurred()))
		concreteRepo, ok := typedRepo.(*genericocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(concreteRepo.ComponentNameMapping).To(Equal(ocmOCIReg.OCIRegistryURLPathMapping))
		Expect(concreteRepo.GetType()).To(Equal(ocireg.Type))
		Expect(concreteRepo.Name()).To(Equal(ociRepoURL))
	})

	It("Then descriptor.component.resources should be correct", func() {
		Expect(len(descriptor.Resources)).To(Equal(1))
		resource := descriptor.Resources[0]
		Expect(resource.Name).To(Equal(module.RawManifestLayerName))
		Expect(resource.Relation).To(Equal(ocmMetaV1.LocalRelation))
		Expect(resource.Type).To(Equal(module.TypeYaml))
		Expect(resource.Version).To(Equal(moduleTemplateVersion))

		resourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
		Expect(err).To(Not(HaveOccurred()))
		localblobAccessSpec, ok := resourceAccessSpec.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(localblobAccessSpec.GetType()).To(Equal(localblob.Type))
		Expect(localblobAccessSpec.LocalReference).To(HavePrefix("sha256:"))
	})

	It("Then descriptor.component.sources should be correct", func() {
		Expect(len(descriptor.Sources)).To(Equal(1))
		source := descriptor.Sources[0]
		sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
		Expect(err).To(Not(HaveOccurred()))
		githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(githubAccessSpec.Type).To(Equal(github.Type))
		Expect(testRepoURL).To(ContainSubstring(githubAccessSpec.RepoURL))
	})

	It("Then security scan labels should be correct", func() {
		secScanLabels := descriptor.Sources[0].Labels
		var devBranch string
		err = yaml.Unmarshal(secScanLabels[1].Value, &devBranch)
		Expect(err).To(Not(HaveOccurred()))
		Expect(devBranch).To(Equal("main"))

		var rcTag string
		err = yaml.Unmarshal(secScanLabels[2].Value, &rcTag)
		Expect(err).To(Not(HaveOccurred()))
		Expect(rcTag).To(Equal("0.5.0"))

		var language string
		err = yaml.Unmarshal(secScanLabels[3].Value, &language)
		Expect(err).To(Not(HaveOccurred()))
		Expect(language).To(Equal("golang-mod"))

		var exclude string
		err = yaml.Unmarshal(secScanLabels[4].Value, &exclude)
		Expect(err).To(Not(HaveOccurred()))
		Expect(exclude).To(Equal("**/test/**,**/*_test.go"))
	})
})
