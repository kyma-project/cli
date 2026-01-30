package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	modulesfake "github.com/kyma-project/cli.v3/internal/modulesv2/fake"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	testExternalCommunityModuleTemplate = kyma.ModuleTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ext-test-module-0.0.3",
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "ext-test-module",
			Version:    "0.0.3",
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
		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		result, err := repo.ListCore(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, "failed to list module templates: test-error", err.Error())
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
		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		result, err := repo.ListCore(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 1)

		require.Equal(t, "test-module", result[0].ModuleName)
		require.Equal(t, "regular", result[0].Channel)
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
		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		result, err := repo.ListLocalCommunity(context.Background())

		require.Len(t, result, 0)
		require.Error(t, err)
		require.Equal(t, "failed to list module templates: test-error", err.Error())
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

		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		result, err := repo.ListLocalCommunity(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "test-module", result[0].ModuleName)
		require.Equal(t, "test-module-namespace/test-module-2", result[0].GetNamespacedName())
	})
}

func TestModuleTemplateRepository_GetLocalCommunity(t *testing.T) {
	t.Run("gets local community module template", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{
			ReturnErr:            nil,
			ReturnModuleTemplate: testCommunityModuleTemplate,
		}

		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		result, err := repo.GetLocalCommunity(
			context.Background(),
			testCommunityModuleTemplate.Spec.ModuleName,
			testCommunityModuleTemplate.Namespace,
		)

		require.NoError(t, err)
		require.Equal(t, "test-module", result.ModuleName)
		require.Equal(t, "test-module-namespace/test-module-2", result.GetNamespacedName())
	})
}

func TestModuleTemplateRepository_SaveCommunityModule(t *testing.T) {
	t.Run("saves external community module template", func(t *testing.T) {
		fakeKymaClient := fake.KymaClient{}
		fakeRootlessDynamic := fake.RootlessDynamicClient{ReturnErr: nil}

		fakeKubeClient := fake.KubeClient{
			TestKymaInterface:            &fakeKymaClient,
			TestRootlessDynamicInterface: &fakeRootlessDynamic,
		}

		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     nil,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

		err := repo.SaveCommunityModule(
			context.Background(),
			modulesfake.ExternalModuleTemplate(nil),
		)

		require.NoError(t, err)
	})
}

func TestModuleTemplatesRepository_ListExternalCommunity(t *testing.T) {
	t.Run("failed to list module templates", func(t *testing.T) {
		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{},
			Err:     errors.New("failed to list external modules"),
		}
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)
		result, err := repo.ListExternalCommunity(context.Background(), []string{"https://irrelevant.url"}, nil)
		require.Error(t, err)
		require.Len(t, result, 0)
		require.Equal(t, "failed to list external modules", err.Error())
	})

	t.Run("lists filtered external community module templates", func(t *testing.T) {
		fakeExternalCommunityRepository := &modulesfake.ExternalModuleTemplatesRepository{
			Modules: []kyma.ModuleTemplate{testExternalCommunityModuleTemplate},
			Err:     nil,
		}
		fakeKymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{},
			},
		}
		fakeKubeClient := fake.KubeClient{
			TestKymaInterface: &fakeKymaClient,
		}

		testData := []struct {
			filter             func(emt *entities.ExternalModuleTemplate) bool
			expectedResultSize int
			moduleName         string
		}{
			{
				filter:             nil,
				expectedResultSize: 1,
				moduleName:         "ext-test-module",
			},

			{
				filter: func(emt *entities.ExternalModuleTemplate) bool {
					return emt.ModuleName == testExternalCommunityModuleTemplate.Spec.ModuleName
				},
				expectedResultSize: 1,
				moduleName:         "ext-test-module",
			},
			{
				filter: func(emt *entities.ExternalModuleTemplate) bool {
					return emt.ModuleName == "invalid-name"
				},
				expectedResultSize: 0,
				moduleName:         "",
			},
		}

		for _, tt := range testData {
			repo := repository.NewModuleTemplatesRepository(&fakeKubeClient, fakeExternalCommunityRepository)

			result, err := repo.ListExternalCommunity(context.Background(), []string{"https://irrelevant.url"}, tt.filter)
			require.NoError(t, err)
			require.Len(t, result, tt.expectedResultSize)
			if len(result) > 0 {
				require.Equal(t, "ext-test-module", result[0].ModuleName)
			}
		}
	})
}
