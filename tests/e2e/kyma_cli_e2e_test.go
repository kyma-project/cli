package e2e

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kyma Deployment, Enabling and Disabling", Ordered, func() {
	It("Should execute kyma alpha deploy command", func() {
		deployCmd := exec.Command("kyma", "alpha", "deploy")
		deployOut, err := deployCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println(string(deployOut))
	})
})
