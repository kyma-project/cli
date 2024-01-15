package scaffold_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCreateScaffold(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CreateScaffold Suite")
}
