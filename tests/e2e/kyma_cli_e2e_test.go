package e2e_test

import (
	"github.com/kyma-project/cli/internal/cli"
	. "github.com/kyma-project/cli/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling", Ordered, func() {
	kcpSystemNamespace := "kcp-system"

	It("Then should install Kyma and Lifecycle Manager successfully", func() {
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

	It("Then should enable template-operator successfully", func() {
		By("Applying the template-operator ModuleTemplate")
		Expect(ApplyModuleTemplate("moduletemplate_template_operator_regular.yaml")).To(Succeed())

		By("Enabling template-operator on Kyma")
		Expect(EnableModuleOnKyma("template-operator")).To(Succeed())

		Eventually(IsModuleReadyInKymaStatus).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault, "template-operator").
			Should(BeTrue())

		By("Then template-operator resources are deployed in the cluster")
		Eventually(IsCRDAvailable).
			WithContext(ctx).
			WithArguments(k8sClient, "samples.operator.kyma-project.io").
			Should(BeTrue())
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "template-operator-controller-manager", "template-operator-system").
			Should(BeTrue())
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "sample-redis-deployment", "manifest-redis").
			Should(BeTrue())
		Eventually(IsCRReady).
			WithContext(ctx).
			WithArguments("sample", "sample-yaml", "kyma-system").
			Should(BeTrue())
	})

	It("Then should disable template-operator successfully", func() {
		By("Disabling template-operator on Kyma")
		Expect(DisableModuleOnKyma("template-operator")).To(Succeed())

		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		By("Then template-operator resources are removed from the cluster")
		Eventually(IsCRDAvailable).
			WithContext(ctx).
			WithArguments(k8sClient, "samples.operator.kyma-project.io").
			Should(BeFalse())
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "template-operator-controller-manager", "template-operator-system").
			Should(BeFalse())
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "sample-redis-deployment", "manifest-redis").
			Should(BeFalse())
		Eventually(IsCRReady).
			WithContext(ctx).
			WithArguments("sample", "sample-yaml", "kyma-system").
			Should(BeFalse())
	})
})
