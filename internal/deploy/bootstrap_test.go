package deploy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/fake"
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
	kcpKymaMock := &mocks.KymaKube{}
	kcpManifestFileBytes, _ := os.ReadFile("../testdata/manifestInKcp.yaml")
	manifestFileBytes, _ := os.ReadFile("../testdata/manifest.yaml")

	manifestObjs, _ := parseManifests(manifestFileBytes)
	kcpManifestObjs, _ := parseManifests(kcpManifestFileBytes)

	kcpMockDeployment := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lifecycle-manager-controller-manager",
				Namespace: "kcp-system",
			},
			Spec: v1.DeploymentSpec{
				Template: v12.PodTemplateSpec{
					Spec: v12.PodSpec{
						Containers: []v12.Container{
							{
								Args: []string{
									"--leader-elect",
									"--enable-webhooks=true",
								},
							},
						},
					},
				},
			},
		},
	)
	kcpKymaMock.On("Static").Return(kcpMockDeployment).Once()

	kymaMock := &mocks.KymaKube{}
	mockDeployment := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lifecycle-manager-controller-manager",
				Namespace: "kcp-system",
			},
			Spec: v1.DeploymentSpec{
				Template: v12.PodTemplateSpec{
					Spec: v12.PodSpec{
						Containers: []v12.Container{
							{
								Args: []string{
									"--leader-elect",
									"--enable-webhooks=true",
								},
							},
						},
					},
				},
			},
		},
	)
	kymaMock.On("Static").Return(mockDeployment).Once()

	type args struct {
		ctx          context.Context
		k8s          kube.KymaKube
		manifestObjs []ctrlClient.Object
		isInKcpMode  bool
	}

	tests := []struct {
		name    string
		args    args
		want    *v1.Deployment
		wantErr bool
	}{
		{
			name: "isInKcpMode is true with no in-kcp-mode flag",
			args: args{
				isInKcpMode:  true,
				ctx:          context.TODO(),
				k8s:          kymaMock,
				manifestObjs: manifestObjs,
			},
			wantErr: false,
			want: &v1.Deployment{
				Spec: v1.DeploymentSpec{
					Template: v12.PodTemplateSpec{
						Spec: v12.PodSpec{
							Containers: []v12.Container{{
								Args: []string{
									"--leader-elect",
									"--enable-webhooks=true",
									"--in-kcp-mode",
								},
							}},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lifecycle-manager-controller-manager",
					Namespace: "kcp-system",
				},
			},
		},
		{
			name: "isInKcpMode is false with in-kcp-mode flag",
			args: args{
				isInKcpMode:  false,
				ctx:          context.TODO(),
				k8s:          kcpKymaMock,
				manifestObjs: kcpManifestObjs,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "isInKcpMode is false with no in-kcp-mode flag",
			args: args{
				isInKcpMode:  false,
				ctx:          context.TODO(),
				k8s:          kymaMock,
				manifestObjs: manifestObjs,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "isInKcpMode is true with in-kcp-mode flag",
			args: args{
				isInKcpMode:  true,
				ctx:          context.TODO(),
				k8s:          kcpKymaMock,
				manifestObjs: kcpManifestObjs,
			},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PatchDeploymentWithInKcpModeFlag(tt.args.ctx, tt.args.k8s, tt.args.manifestObjs, tt.args.isInKcpMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchDeploymentWithInKcpMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PatchDeploymentWithInKcpModeFlag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func parseManifests(manifest []byte) ([]ctrlClient.Object, error) {
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
