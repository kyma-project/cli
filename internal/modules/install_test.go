package modules

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getModuleTemplateSpecWithResourceLink(link string) kyma.ModuleTemplate {
	return kyma.ModuleTemplate{
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "serverless",
			Version:    "0.0.1",
			Data:       unstructured.Unstructured{},
			Resources: []kyma.Resource{
				{
					Name: "rawManifest",
					Link: link,
				},
			},
		},
	}
}

var validResourceYaml = `---
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-ip-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-ip-system-2
---
`

var invalidResourceYaml = `apiVersion: v1
		kind: Namespace
	metadata:
  name: cluster-ip-system
`

func TestInstall_ListModuleTemplateError(t *testing.T) {
	ctx := context.Background()
	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{},
		ReturnErr:                errors.New("ListModuleTemplateError"),
	}
	client := fake.KubeClient{
		TestKymaInterface: &kymaClient,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: nil,
		IsDefaultCRApplicable:   false,
	}
	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCore: []kyma.ModuleTemplate{},
	}

	expectedCliErr := clierror.New("cannot install non-existing module")

	clierr := Install(ctx, &client, repo, data)
	require.NotNil(t, clierr)
	require.Equal(t, expectedCliErr, clierr)
}

func TestInstall_InstallModuleError(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(invalidResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	client := fake.KubeClient{}

	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunityByName: []kyma.ModuleTemplate{testModuleTemplate},
		CommunityByNameErr:    nil,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: &testModuleTemplate,
		IsDefaultCRApplicable:   false,
	}

	expectedCliErr := clierror.Wrap(
		errors.New("failed to apply resources from link: "+
			"failed to parse module resource: yaml: "+
			"line 2: found a tab character that violates indentation"),
		clierror.New("failed to install community module"),
	)

	clierr := Install(ctx, &client, repo, data)
	require.NotNil(t, clierr)
	require.Equal(t, expectedCliErr, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledFromRemote(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunityByName: []kyma.ModuleTemplate{testModuleTemplate},
		CommunityByNameErr:    nil,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: &testModuleTemplate,
		IsDefaultCRApplicable:   false,
	}

	clierr := Install(ctx, &client, repo, data)
	require.Nil(t, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledFromLocal(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	rootlessDynamicClient := fake.RootlessDynamicClient{}

	fakeKyma := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{testModuleTemplate},
		},
	}

	client := fake.KubeClient{
		TestKymaInterface:            &fakeKyma,
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunityByName: []kyma.ModuleTemplate{},
		CommunityByNameErr:    nil,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: &testModuleTemplate,
		IsDefaultCRApplicable:   false,
	}

	clierr := Install(ctx, &client, repo, data)
	require.Nil(t, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledWithDefaultCR(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunityByName: []kyma.ModuleTemplate{testModuleTemplate},
		CommunityByNameErr:    nil,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: &testModuleTemplate,
		IsDefaultCRApplicable:   true,
	}

	clierr := Install(ctx, &client, repo, data)
	require.Nil(t, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledWithCustomCR(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	repo := &modulesfake.ModuleTemplatesRepo{
		ReturnCommunityByName: []kyma.ModuleTemplate{testModuleTemplate},
		CommunityByNameErr:    nil,
	}

	data := InstallCommunityModuleData{
		CommunityModuleTemplate: &testModuleTemplate,
		IsDefaultCRApplicable:   false,
		CustomResources:         []unstructured.Unstructured{},
	}

	clierr := Install(ctx, &client, repo, data)
	require.Nil(t, clierr)
}

func TestFindCommunityModuleTemplate(t *testing.T) {
	ctx := context.Background()
	namespace := "my-system"
	communityModuleTemplateName := "community-module-0.1.0"

	t.Run("module not found", func(t *testing.T) {
		repo := modulesfake.ModuleTemplatesRepo{
			ReturnCommunity: []kyma.ModuleTemplate{},
		}
		foundModule, err := FindCommunityModuleTemplate(ctx, namespace, communityModuleTemplateName, &repo)

		require.Error(t, err)
		require.Nil(t, foundModule)
		require.Contains(t, err.Error(), "module of the provided origin does not exist")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := modulesfake.ModuleTemplatesRepo{
			CommunityErr: errors.New("repo error"),
		}
		foundModule, err := FindCommunityModuleTemplate(ctx, namespace, communityModuleTemplateName, &repo)

		require.Error(t, err)
		require.Nil(t, foundModule)
		require.Contains(t, err.Error(), "failed to retrieve community modules: repo error")
	})
}

func getTestHttpServerWithResponse(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
		w.WriteHeader(http.StatusOK)
	}))
}
