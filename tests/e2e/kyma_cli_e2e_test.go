package e2e_test

import (
	"os/exec"

	. "github.com/kyma-project/cli/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling", Ordered, func() {
	kubeConfig := ctrl.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(kubeConfig)
	kcpSystemNamespace := "kcp-system"

	It("Then should install Kyma and Lifecycle Manager successfully", func() {
		By("Executing kyma alpha deploy command")
		deployCmd := exec.Command("kyma", "alpha", "deploy")
		deployOut, err := deployCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		Eventually(deployOut).Should(ContainSubstring("Happy Kyma-ing!"))

		//By("Then Kyma CR should be deployed")
		//kymaResource := schema.GroupVersionResource{
		//	Group:    "operator.kyma-project.io",
		//	Version:  "v1beta2",
		//	Resource: "kymas",
		//}

		By("Then Lifecycle Manager should be Ready")
		Eventually(IsDeploymentReady).
			WithContext(ctx).
			WithArguments(client, "lifecycle-manager-controller-manager", kcpSystemNamespace).
			Should(BeTrue())

	})
})
