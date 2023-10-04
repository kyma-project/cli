package e2e_test

import (
	"github.com/kyma-project/cli/internal/cli"
	. "github.com/kyma-project/cli/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling Module", Ordered, func() {
	kcpSystemNamespace := "kcp-system"
	deployments := map[string]string{
		"template-operator-controller-manager": "template-operator-system",
		"sample-redis-deployment":              "manifest-redis",
	}
	BeforeAll(func() {
		By("Executing kyma alpha deploy command")
		Expect(ExecuteKymaDeployCommand()).To(Succeed())

		By("Then Kyma CR should be Ready")
		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		By("And Lifecycle Manager should be Ready")
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "lifecycle-manager-controller-manager", kcpSystemNamespace).
			Should(BeTrue())
	})

	Context("Enabling and disabling ready state template-operator successfully", func() {
		It("Enabling a ready state module", func() {
			By("Applying the template-operator ModuleTemplate")
			Expect(ApplyModuleTemplate("module_templates/moduletemplate_template_operator_regular.yaml")).
				To(Succeed())

			By("Enabling template-operator on Kyma")
			Expect(EnableModuleOnKymaWithReadyStateModule("template-operator")).To(Succeed())

			Eventually(IsModuleReadyInKymaStatus).
				WithContext(ctx).
				WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault, "template-operator").
				Should(BeTrue())

			By("Then template-operator resources are deployed in the cluster")
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeTrue())
			Eventually(IsCRReady).
				WithContext(ctx).
				WithArguments("sample", "sample-yaml", "kyma-system").
				Should(BeTrue())
		})

		It("Disabling a ready state template-operator on Kyma", func() {
			By("Executing kyma disable module command")
			Expect(DisableModuleOnKyma("template-operator")).To(Succeed())

			Eventually(IsKymaCRInReadyState).
				WithContext(ctx).
				WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
				Should(BeTrue())

			By("Then template-operator resources are removed from the cluster")
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeFalse())
		})

	})

	Context("Enabling and disabling warning state template-operator successfully", func() {
		It("Enabling a warning state module", func() {
			By("Applying the template-operator ModuleTemplate")
			Expect(ApplyModuleTemplate(
				"module_templates/moduletemplate_template_operator_regular_warning.yaml")).
				To(Succeed())

			By("Enabling template-operator on Kyma")
			Expect(EnableModuleOnKymaWithWarningStateModule("template-operator")).To(Succeed())

			Eventually(IsModuleInWarningStateInKymaStatus).
				WithContext(ctx).
				WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault, "template-operator").
				Should(BeTrue())

			By("Then template-operator resources are deployed in the cluster")
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeTrue())

			Eventually(IsCRInWarningState).
				WithContext(ctx).
				WithArguments("sample", "sample-yaml", "kyma-system").
				Should(BeTrue())
		})

		It("Disabling a warning state template-operator on Kyma", func() {
			By("Executing kyma disable module command")
			Expect(DisableModuleOnKyma("template-operator")).To(Succeed())

			Eventually(IsKymaCRInReadyState).
				WithContext(ctx).
				WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
				Should(BeTrue())

			By("Then template-operator resources are removed from the cluster")
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io",
					deployments,
					"sample", "sample-yaml", "kyma-system").
				Should(BeFalse())
		})
	})
})
