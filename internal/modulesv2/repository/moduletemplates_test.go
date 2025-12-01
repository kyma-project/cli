package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testCoreModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-1",
			Namespace: "kyma-system",
			Labels: map[string]string{
				"operator.kyma-project.io/managed-by": "kyma",
			},
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "0.0.1",
		},
	}
	testLegacyModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-1",
			Namespace: "kyma-system",
			Labels: map[string]string{
				"operator.kyma-project.io/managed-by": "kyma",
			},
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "0.0.1",
			Channel:    "fast",
			Descriptor: runtime.RawExtension{
				Raw: []byte(`component:
  name:test-module
  version:0.0.1`),
			},
		},
	}
	testCoreModuleTemplateReleaseMeta = kyma.ModuleReleaseMeta{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleReleaseMeta",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-1",
			Namespace: "kyma-system",
		},
		Spec: kyma.ModuleReleaseMetaSpec{
			ModuleName: "test-module",
			Channels: []kyma.ChannelVersionAssignment{
				{
					Channel: "regular",
					Version: "0.0.1",
				},
				{
					Channel: "fast",
					Version: "0.0.2",
				},
				{
					Channel: "experimental",
					Version: "0.0.3",
				},
			},
		},
	}
	testCommunityModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-module-2",
			Namespace: "test-module-namespace",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "test-module",
			Version:    "0.0.2",
		},
	}
)

func TestModuleTemplatesRepository_ListCore(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: errors.New("test-error"),
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}
		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, repository.NewExternalModuleTemplateRepository())

		result, err := repo.ListCore(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, err.Error(), "failed to list module templates: test-error")
	})

	t.Run("lists all core module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{
					testCoreModuleTemplate,
					testCommunityModuleTemplate,
				},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{
					testCoreModuleTemplateReleaseMeta,
				},
			},
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, repository.NewExternalModuleTemplateRepository())

		result, err := repo.ListCore(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 3)

		require.Equal(t, "test-module", result[0].ModuleName)
		require.Equal(t, "regular", result[0].Channel)

		require.Equal(t, "test-module", result[1].ModuleName)
		require.Equal(t, "fast", result[1].Channel)

		require.Equal(t, "test-module", result[2].ModuleName)
		require.Equal(t, "experimental", result[2].Channel)
	})
}

func TestModuleTemplatesRepository_ListLocalCommunity(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: errors.New("test-error"),
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}
		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, repository.NewExternalModuleTemplateRepository())

		result, err := repo.ListLocalCommunity(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, err.Error(), "failed to list module templates: test-error")
	})

	t.Run("lists all core module templates", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{
					testCoreModuleTemplate,
					testCommunityModuleTemplate,
				},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{
					testCoreModuleTemplateReleaseMeta,
				},
			},
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, repository.NewExternalModuleTemplateRepository())

		result, err := repo.ListLocalCommunity(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "test-module", result[0].ModuleName)
		require.Equal(t, "test-module-namespace/test-module-2", result[0].GetNamespacedName())
	})
}

func TestModuleTemplatesRepository_ListExternalCommunity(t *testing.T) {
}
