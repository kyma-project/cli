package e2e_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling", Ordered, func() {
	kubeConfig := ctrl.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(kubeConfig)
	kcpSystemNamespace := "kcp-system"

	It("Then should install Kyma and Lifecycle Manager successfully", func() {
		When("Executing kyma alpha deploy command")
		deployCmd := exec.Command("kyma", "alpha", "deploy")
		deployOut, err := deployCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		Eventually(deployOut).Should(ContainSubstring("Happy Kyma-ing!"))

		By("Then Kyma CR should be deployed")

		By("Then Lifecycle Manager should be deployed")
		lifecycleManager, err := client.AppsV1().Deployments(kcpSystemNamespace).
			Get(ctx, "lifecycle-manager-controller-manager", v1.GetOptions{})
		GinkgoWriter.Println(lifecycleManager.Name)
	})
})
