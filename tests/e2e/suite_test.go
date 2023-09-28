package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = BeforeSuite(func() {
	go func() {
		defer GinkgoRecover()
	}()
})
