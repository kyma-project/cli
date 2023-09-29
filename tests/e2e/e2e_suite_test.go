package e2e_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout  = 2 * time.Minute
	interval = 1 * time.Second
)

var (
	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient client.Client
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())
	SetDefaultEventuallyPollingInterval(interval)
	SetDefaultEventuallyTimeout(timeout)

	kubeConfig := ctrl.GetConfigOrDie()
	Expect(kubeConfig).NotTo(BeNil())
	var err error
	//Expect(v1beta2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	k8sClient, err = client.New(kubeConfig, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	go func() {
		defer GinkgoRecover()
	}()
})

var _ = AfterSuite(func() {
	cancel()
})
