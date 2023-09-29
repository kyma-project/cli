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
		When("Executing kyma alpha deploy command", func() {
			deployCmd := exec.Command("kyma", "alpha", "deploy")
			deployOut, err := deployCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Eventually(deployOut).Should(ContainSubstring("Happy Kyma-ing!"))
		})

		By("Then Kyma CR should be deployed", func() {

		})

		By("Then Lifecycle Manager should be deployed", func() {
			lifecycleManager, err := client.AppsV1().Deployments(kcpSystemNamespace).
				Get(ctx, "lifecycle-manager-controller-manager", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			GinkgoWriter.Println(lifecycleManager.Status)
		})

	})
})
