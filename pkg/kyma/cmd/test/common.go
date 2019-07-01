package test

import (
	"strings"

	"github.com/kyma-project/cli/internal/kubectl"
)

func ListTestSuiteNames(kClient *kubectl.Wrapper) ([]string, error) {
	res, err := kClient.RunCmd("-n", "kyma-system", "get", "clustertestsuites.testing.kyma-project.io", "-o", "custom-columns=:.metadata.name")

	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(res), " "), nil
}

func ListTestDefinitionNames(kClient *kubectl.Wrapper) ([]string, error) {
	res, err := kClient.RunCmd("-n", "kyma-system", "get", "testdefinitions.testing.kyma-project.io", "-o", "custom-columns=:.metadata.name")

	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(res), " "), nil
}
