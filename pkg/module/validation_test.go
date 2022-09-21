package module

import (
	"strings"
	"testing"
)

func TestEnsureDefaultNamespace(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		shouldErr          bool
		outputVal          string
		errorShouldContain string
	}{
		{
			name:      "happy path",
			input:     correctModel,
			shouldErr: false,
			outputVal: correctModel,
		},
		{
			name:               "invalid YAML",
			input:              invalidYaml,
			shouldErr:          true,
			errorShouldContain: InvalidYamlError,
		},
		{
			name:               "missing metadata",
			input:              noMetadataYaml,
			shouldErr:          true,
			errorShouldContain: noMetadataError,
		},
		{
			name:               "metadata is not map",
			input:              invalidMetadataTypeYaml,
			shouldErr:          true,
			errorShouldContain: invalidMetadataTypeError,
		},
		{
			name:               "metadata.namespace is not string",
			input:              invalidNamespaceTypeYaml,
			shouldErr:          true,
			errorShouldContain: invalidNamespaceTypeError,
		},
		{
			name:      "should add metadata.namespace if missing",
			input:     missingNamespaceYaml,
			shouldErr: false,
			outputVal: missingNamespaceExpectedOutput,
		},
		{
			name:      "should set metadata.namespace to default if different",
			input:     differentNamespaceYaml,
			shouldErr: false,
			outputVal: differentNamespaceExpectedOutput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ensureDefaultNamespace([]byte(tc.input))
			if err == nil && tc.shouldErr {
				t.Errorf("ensureDefaultNamespace() should return an error but it didn't")
			}
			if err != nil && !tc.shouldErr {
				t.Errorf("ensureDefaultNamespace() should not return an error but it did: %v", err)
			}
			if err != nil && !strings.Contains(err.Error(), tc.errorShouldContain) {
				t.Errorf("ensureDefaultNamespace() should return an error containing: \"%v\"\nbut it returned: %v", tc.errorShouldContain, err.Error())
			}
			if err == nil && string(output) != tc.outputVal {
				t.Errorf("ensureDefaultNamespace() should return:\n%v\nbut it returned:\n%v", tc.outputVal, string(output))
			}
		})

	}
}

const (
	correctModel = `apiVersion: operator.kyma-project.io/v1alpha1
kind: Sample
metadata:
    name: sample-sample
    namespace: default
spec:
    releaseName: redis-release
`
	invalidYaml      = `!abc: def`
	InvalidYamlError = "cannot unmarshal !abc: `def` into map[string]interface {}"

	noMetadataYaml = `abc: def
foo: bar
`
	noMetadataError = `no "metadata" attribute`

	invalidMetadataTypeYaml = `
abc: def
metadata: 2
`
	invalidMetadataTypeError = `"metadata" attribute is not a Map`

	invalidNamespaceTypeYaml = `
abc: def
metadata:
  namespace: 2
`
	invalidNamespaceTypeError = `"metadata.namespace" attribute is not a string`
	missingNamespaceYaml      = `abc: def
metadata:
    name: foobar
`
	missingNamespaceExpectedOutput = `abc: def
metadata:
    name: foobar
    namespace: default
`

	differentNamespaceYaml = `abc: def
metadata:
    name: foobar
    namespace: foobar
`
	differentNamespaceExpectedOutput = `abc: def
metadata:
    name: foobar
    namespace: default
`
)
