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

	Context("Given a Kyma Cluster", func() {
		It("When `kyma alpha deploy` command is executed")
		Expect(ExecuteKymaDeployCommand()).To(Succeed())

		By("Then the Kyma CR is in a ready state")
		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		By("And the Lifecycle Manager is in a ready state")
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "lifecycle-manager-controller-manager", kcpSystemNamespace).
			Should(BeTrue())
	})

	Context("Given a valid Template Operator module template", func() {
		It("When a Template Operator module is applied", func() {
			Expect(ApplyModuleTemplate("module_templates/moduletemplate_template_operator_regular.yaml")).
				To(Succeed())
		})

		By("And the Template Operator gets enabled")
		Expect(EnableModuleOnKymaWithReadyStateModule("template-operator")).To(Succeed())

		Eventually(IsModuleReadyInKymaStatus).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault, "template-operator").
			Should(BeTrue())

		It("Then Template Operator resources are deployed in the cluster", func() {
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeTrue())
		})

		It("And the Template Operator's CR state is ready", func() {
			Eventually(IsCRReady).
				WithContext(ctx).
				WithArguments("sample", "sample-yaml", "kyma-system").
				Should(BeTrue())
		})
	})

	Context("Given a Template Operator module in a ready state", func() {
		It("When `kyma disable module` command is execute", func() {
			Expect(DisableModuleOnKyma("template-operator")).To(Succeed())
		})

		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		It("Then the Template Operator's resources are removed from the cluster", func() {
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeFalse())
		})
	})

	Context("Given a warning state Template Operator module template", func() {
		It("When a Template Operator module is applied", func() {
			Expect(ApplyModuleTemplate(
				"module_templates/moduletemplate_template_operator_regular_warning.yaml")).
				To(Succeed())
		})

		It("And the Template Operator enable command invoked", func() {
			Expect(EnableModuleOnKymaWithWarningStateModule("template-operator")).To(Succeed())
		})

		Eventually(IsModuleInWarningStateInKymaStatus).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault, "template-operator").
			Should(BeTrue())

		It("Then the Template Operator's resources are deployed in the cluster", func() {
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeTrue())
		})

		It("And the Template Operator's CR state is in a warning state", func() {
			Eventually(IsCRInWarningState).
				WithContext(ctx).
				WithArguments("sample", "sample-yaml", "kyma-system").
				Should(BeTrue())
		})
	})

	Context("Given a Template Operator module in a warning state", func() {
		It("When `kyma disable module` command is executed", func() {
			Expect(DisableModuleOnKyma("template-operator")).To(Succeed())
		})

		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		It("Then Template Operator's resources are removed from the cluster", func() {
			Eventually(AreModuleResourcesReadyInCluster).
				WithContext(ctx).
				WithArguments(k8sClient, "samples.operator.kyma-project.io", deployments).
				Should(BeFalse())
		})
	})
})
