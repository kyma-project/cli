package e2e_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	timeout  = 10 * time.Second
	interval = 1 * time.Second
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyPollingInterval(interval)
	SetDefaultEventuallyTimeout(timeout)
})
