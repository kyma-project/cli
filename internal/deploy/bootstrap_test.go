package deploy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

func TestPatchDeploymentWithInKcpModeFlag(t *testing.T) {
	kcpManifestFileBytes, _ := os.ReadFile("../testdata/manifestInKcp.yaml")
	manifestFileBytes, _ := os.ReadFile("../testdata/manifest.yaml")

	manifestObjs, _ := parseManifestsBytes(manifestFileBytes)
	kcpManifestObjs, _ := parseManifestsBytes(kcpManifestFileBytes)

	type args struct {
		manifestObjs []ctrlClient.Object
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "non-existing in-kcp-mode flag",
			args: args{
				manifestObjs: manifestObjs,
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "existing in-kcp-mode flag",
			args: args{
				manifestObjs: kcpManifestObjs,
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := patchDeploymentWithInKcpModeFlag(tt.args.manifestObjs)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchDeploymentWithInKcpMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			hasKcpFlag := validateDeploymentHasKcpFlag(t, tt.args.manifestObjs)
			if !assert.Equal(t, tt.want, hasKcpFlag) {
				t.Errorf("PatchDeploymentWithInKcpModeFlag() got = %t, want %t", hasKcpFlag, tt.want)
			}
		})
	}
}

func validateDeploymentHasKcpFlag(t *testing.T, objs []ctrlClient.Object) bool {
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
			unstr, ok := obj.(*unstructured.Unstructured)
			assert.True(t, ok)
			manifestJSON, err := json.Marshal(unstr.Object)
			assert.NoError(t, err)
			deployment := &appsv1.Deployment{}
			if err := json.Unmarshal(manifestJSON, deployment); err == nil {
				if slices.Contains(deployment.Spec.Template.Spec.Containers[0].Args, "--in-kcp-mode") {
					return true
				}
			}
		}
	}
	return false
}

func parseManifestsBytes(manifest []byte) ([]ctrlClient.Object, error) {
	manifests, err := parseManifest(manifest)
	if err != nil {
		return nil, err
	}

	objs := make([]ctrlClient.Object, 0, len(manifests))
	for _, manifest := range manifests {
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(manifest, obj); err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}

	return objs, nil
}

func parseManifest(data []byte) ([][]byte, error) {
	var chanBytes [][]byte
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))

	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				return chanBytes, nil
			}
			return nil, err
		}
		chanBytes = append(chanBytes, buf)
	}
}
