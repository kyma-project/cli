package e2e_test

import (
	"os/exec"

	"github.com/kyma-project/cli/internal/cli"
	. "github.com/kyma-project/cli/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling", Ordered, func() {
	kcpSystemNamespace := "kcp-system"

	It("Then should install Kyma and Lifecycle Manager successfully", func() {
		By("Executing kyma alpha deploy command")
		deployCmd := exec.Command("kyma", "alpha", "deploy")
		deployOut, err := deployCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		Eventually(string(deployOut)).Should(ContainSubstring("Kyma CR deployed and Ready"))

		By("Then Kyma CR should be Ready")
		Eventually(IsKymaCRInReadyState).
			WithContext(ctx).
			WithArguments(k8sClient, cli.KymaNameDefault, cli.KymaNamespaceDefault).
			Should(BeTrue())

		By("Then Lifecycle Manager should be Ready")
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "lifecycle-manager-controller-manager", kcpSystemNamespace).
			Should(BeTrue())
	})

	It("Then should enable template-operator successfully", func() {
		By("Applying the template-operator ModuleTemplate")

		By("Then template-operator ModuleTemplate should be available in cluster")

		By("Then Lifecycle Manager should be Ready")
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(k8sClient, "lifecycle-manager-controller-manager", kcpSystemNamespace).
			Should(BeTrue())
	})
})
