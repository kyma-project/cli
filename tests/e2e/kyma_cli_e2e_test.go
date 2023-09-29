package e2e_test

import (
	"fmt"
	"os"

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
		templateOperatorModuleTemplate := os.Getenv("MODULE_TEMPLATE_PATH")
		GinkgoWriter.Println(templateOperatorModuleTemplate)
		Expect(ApplyModuleTemplate(fmt.Sprintf("%s/%s", "../../template-operator/",
			templateOperatorModuleTemplate))).Should(Succeed())

		By("Enabling template-operator on Kyma")

		By("Then template-operator resources are deployed in the cluster")

	})
})
