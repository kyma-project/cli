package deploy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasKyma(t *testing.T) {
	manifest := `apiVersion: apiextensions.k8s.io/v1
	kind: CustomResourceDefinition
	metadata:
	  annotations:
		controller-gen.kubebuilder.io/version: v0.9.2
	  creationTimestamp: null
	  labels:
		app.kubernetes.io/component: lifecycle-manager.kyma-project.io
		app.kubernetes.io/created-by: kustomize
		app.kubernetes.io/instance: kcp-lifecycle-manager-main
		app.kubernetes.io/managed-by: kustomize
		app.kubernetes.io/name: kcp-lifecycle-manager
		app.kubernetes.io/part-of: manual-deployment
	  name: kymas.operator.kyma-project.io
	spec:
	  group: operator.kyma-project.io
	  names:
		kind: Kyma
		listKind: KymaList
		plural: kymas
		singular: kyma
	  scope: Namespaced
	  versions:
	  - additionalPrinterColumns:
		- jsonPath: .status.state
		  name: State
		  type: string
		- jsonPath: .metadata.creationTimestamp
		  name: Age
		  type: date
		name: v1alpha1
		schema:
		#...
	`

	b, err := hasKyma(manifest)
	require.NoError(t, err)
	require.True(t, b)

}
