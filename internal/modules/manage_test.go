package modules

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

var (
	moduleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-123",
			Namespace: "kyma-system",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "1.2.3",
			Data:       unstructured.Unstructured{},
			Manager: &kyma.Manager{
				Name: "test-module-controller-manager",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
		},
	}

	moduleReleaseMeta = &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]any{
				"name":      "test-module",
				"namespace": "kyma-system",
			},
			"spec": map[string]any{
				"moduleName": "test-module",
				"channels": []any{
					map[string]any{
						"version": "1.2.3",
						"channel": "regular",
					},
					map[string]any{
						"version": "1.2.4",
						"channel": "fast",
					},
				},
			},
		},
	}

	defaultKymaEmpty = &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]any{
				"name":      "default",
				"namespace": "kyma-system",
			},
			"spec": map[string]any{
				"channel": "fast",
				"modules": []any{},
			},
			"status": map[string]any{
				"modules": []any{},
			},
		},
	}

	defaultKyma = &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]any{
				"name":      "default",
				"namespace": "kyma-system",
			},
			"spec": map[string]any{
				"channel": "fast",
				"modules": []any{
					map[string]any{
						"name": "test-module",
					},
				},
			},
			"status": map[string]any{
				"modules": []any{
					map[string]any{
						"name":    "test-module",
						"version": "1.2.3",
						"state":   "Ready",
					},
				},
			},
		},
	}

	manager = unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"namespace": "test-module-system",
				"labels": map[string]any{
					"app.kubernetes.io/version": "1.2.3",
				},
				"name": "test-module-controller-manager3",
			},
			"spec":   map[string]any{},
			"status": map[string]any{},
		},
	}
)

func TestModuleExistsInKymaCR(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRKyma.GroupVersion())
	kymaClient := kyma.NewClient(dynamic_fake.NewSimpleDynamicClient(scheme, defaultKyma))

	client := &fake.KubeClient{
		TestKymaInterface: kymaClient,
	}

	exists, err := ModuleExistsInKymaCR(context.Background(), client, "test-module")
	require.NoError(t, err)
	require.True(t, exists)

	notExists, err := ModuleExistsInKymaCR(context.Background(), client, "non-existing-module")
	require.NoError(t, err)
	require.False(t, notExists)
}

func TestManageModuleInKymaCR_Failed(t *testing.T) {
	kymaClient := fake.KymaClient{
		ReturnWaitForModuleErr: errors.New("test error"),
	}

	client := &fake.KubeClient{
		TestKymaInterface: &kymaClient,
	}

	err := ManageModuleInKymaCR(context.Background(), client, "test-module", "CreateAndDelete")
	require.Error(t, err)
	require.Equal(t, "failed to set module as managed: test error", err.Error())
}

func TestManageModuleInKymaCR_Success(t *testing.T) {
	kymaClient := fake.KymaClient{
		ReturnWaitForModuleErr: nil,
	}

	client := &fake.KubeClient{
		TestKymaInterface: &kymaClient,
	}

	err := ManageModuleInKymaCR(context.Background(), client, "test-module", "CreateAndDelete")
	require.NoError(t, err)
}

func TestManageModuleMissingInKyma_ModuleNotInstalled(t *testing.T) {
	fakeRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCore: []kyma.ModuleTemplate{},
	}

	fakeClient := &fake.KubeClient{}

	err := ManageModuleMissingInKyma(
		context.Background(),
		fakeClient,
		fakeRepo,
		"test-module",
		kyma.CustomResourcePolicyCreateAndDelete,
	)

	require.Error(t, err)
	require.Equal(t, "failed to find installed module", err.Error())
}

func TestManageModuleMissingInKyma_ModuleVersionNotPresentInKymaChannel(t *testing.T) {
	fakeRepo := &modulesfake.ModuleTemplatesRepo{
		ReturnCore:             []kyma.ModuleTemplate{moduleTemplate},
		ReturnInstalledManager: &manager,
	}

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
	fakeKymaClient := kyma.NewClient(dynamic_fake.NewSimpleDynamicClient(scheme, defaultKymaEmpty, moduleReleaseMeta))

	fakeClient := &fake.KubeClient{
		TestKymaInterface: fakeKymaClient,
	}

	err := ManageModuleMissingInKyma(
		context.Background(),
		fakeClient,
		fakeRepo,
		"test-module",
		kyma.CustomResourcePolicyCreateAndDelete,
	)

	require.Error(t, err)
	require.Equal(t, ErrModuleInstalledVersionNotInKymaChannel, err)
}

func TestGetAvailableChannelsAndVersions(t *testing.T) {
	testCases := []struct {
		name        string
		repoCore    []kyma.ModuleTemplate
		repoCoreErr error
		metas       *kyma.ModuleReleaseMetaList
		metasErr    error
		moduleName  string
		expected    map[string]string
		expectErr   bool
	}{
		{
			name:       "single channel/version",
			repoCore:   []kyma.ModuleTemplate{{Spec: kyma.ModuleTemplateSpec{ModuleName: "foo", Version: "1.0.0"}}},
			metas:      &kyma.ModuleReleaseMetaList{Items: []kyma.ModuleReleaseMeta{{Spec: kyma.ModuleReleaseMetaSpec{ModuleName: "foo", Channels: []kyma.ChannelVersionAssignment{{Channel: "stable", Version: "1.0.0"}}}}}},
			moduleName: "foo",
			expected:   map[string]string{"stable": "1.0.0"},
		},
		{
			name: "multiple channels/versions",
			repoCore: []kyma.ModuleTemplate{
				{Spec: kyma.ModuleTemplateSpec{ModuleName: "foo", Version: "1.0.0"}},
				{Spec: kyma.ModuleTemplateSpec{ModuleName: "foo", Version: "2.0.0"}},
			},
			metas:      &kyma.ModuleReleaseMetaList{Items: []kyma.ModuleReleaseMeta{{Spec: kyma.ModuleReleaseMetaSpec{ModuleName: "foo", Channels: []kyma.ChannelVersionAssignment{{Channel: "stable", Version: "1.0.0"}, {Channel: "dev", Version: "2.0.0"}}}}}},
			moduleName: "foo",
			expected:   map[string]string{"stable": "1.0.0", "dev": "2.0.0"},
		},
		{
			name:       "no matching module",
			repoCore:   []kyma.ModuleTemplate{{Spec: kyma.ModuleTemplateSpec{ModuleName: "bar", Version: "1.0.0"}}},
			metas:      &kyma.ModuleReleaseMetaList{Items: []kyma.ModuleReleaseMeta{}},
			moduleName: "foo",
			expected:   map[string]string{},
		},
		{
			name:        "repo.Core error",
			repoCore:    nil,
			repoCoreErr: errors.New("fail core"),
			metas:       &kyma.ModuleReleaseMetaList{},
			moduleName:  "foo",
			expectErr:   true,
		},
		{
			name:       "ListModuleReleaseMeta error",
			repoCore:   []kyma.ModuleTemplate{{Spec: kyma.ModuleTemplateSpec{ModuleName: "foo", Version: "1.0.0"}}},
			metas:      &kyma.ModuleReleaseMetaList{},
			metasErr:   errors.New("fail meta"),
			moduleName: "foo",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &modulesfake.ModuleTemplatesRepo{
				ReturnCore: tc.repoCore,
				CoreErr:    tc.repoCoreErr,
			}
			kymaClient := fake.KymaClient{
				ReturnModuleReleaseMetaList: *tc.metas,
				ReturnErr:                   tc.metasErr,
			}
			fakeClient := &fake.KubeClient{
				TestKymaInterface: &kymaClient,
			}

			result, err := GetAvailableChannelsAndVersions(context.Background(), fakeClient, repo, tc.moduleName)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}
