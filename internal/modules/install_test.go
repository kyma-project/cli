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

var validResourceYaml = `apiVersion: v1
kind: Namespace
metadata:
  name: cluster-ip-system
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
		ModuleName:            "non-existing-module",
		Version:               "v0.0.1",
		IsDefaultCRApplicable: false,
	}

	expectedClieErr := clierror.Wrap(
		errors.New("failed to get module template"),
		clierror.New("failed to retrieve community module from catalog"),
	)

	clierr := Install(ctx, &client, data)
	require.NotNil(t, clierr)
	require.Equal(t, expectedClieErr, clierr)
}

func TestInstall_ModuleNotFound(t *testing.T) {
	ctx := context.Background()
	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{},
		ReturnErr:                nil,
	}
	client := fake.KubeClient{
		TestKymaInterface: &kymaClient,
	}

	data := InstallCommunityModuleData{
		ModuleName:            "non-existing-module",
		Version:               "v0.0.1",
		IsDefaultCRApplicable: false,
	}

	expectedClieErr := clierror.Wrap(
		errors.New("module not found"),
		clierror.New("failed to retrieve community module from catalog"),
	)

	clierr := Install(ctx, &client, data)
	require.NotNil(t, clierr)
	require.Equal(t, expectedClieErr, clierr)
}

func TestInstall_InstallModuleError(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(invalidResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{testModuleTemplate},
		},
		ReturnErr: nil,
	}
	client := fake.KubeClient{
		TestKymaInterface: &kymaClient,
	}

	data := InstallCommunityModuleData{
		ModuleName:            "serverless",
		Version:               "0.0.1",
		IsDefaultCRApplicable: false,
	}

	expectedClieErr := clierror.Wrap(
		errors.New("failed to apply resources from link: "+
			"failed to parse module resource: yaml: "+
			"line 2: found a tab character that violates indentation"),
		clierror.New("failed to install community module"),
	)

	clierr := Install(ctx, &client, data)
	require.NotNil(t, clierr)
	require.Equal(t, expectedClieErr, clierr)
}

func TestInstall_ModuleSuccessfullyInstalled(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{testModuleTemplate},
		},
		ReturnErr: nil,
	}
	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestKymaInterface:            &kymaClient,
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	data := InstallCommunityModuleData{
		ModuleName:            "serverless",
		Version:               "0.0.1",
		IsDefaultCRApplicable: false,
	}

	clierr := Install(ctx, &client, data)
	require.Nil(t, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledWithDefaultCR(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{testModuleTemplate},
		},
		ReturnErr: nil,
	}
	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestKymaInterface:            &kymaClient,
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	data := InstallCommunityModuleData{
		ModuleName:            "serverless",
		Version:               "0.0.1",
		IsDefaultCRApplicable: true,
	}

	clierr := Install(ctx, &client, data)
	require.Nil(t, clierr)
}

func TestInstall_ModuleSuccessfullyInstalledWithCustomCR(t *testing.T) {
	ctx := context.Background()

	testHttpServer := getTestHttpServerWithResponse(validResourceYaml)
	defer testHttpServer.Close()

	testModuleTemplate := getModuleTemplateSpecWithResourceLink(testHttpServer.URL)

	kymaClient := fake.KymaClient{
		ReturnModuleTemplateList: kyma.ModuleTemplateList{
			Items: []kyma.ModuleTemplate{testModuleTemplate},
		},
		ReturnErr: nil,
	}
	rootlessDynamicClient := fake.RootlessDynamicClient{}

	client := fake.KubeClient{
		TestKymaInterface:            &kymaClient,
		TestRootlessDynamicInterface: &rootlessDynamicClient,
	}

	data := InstallCommunityModuleData{
		ModuleName:            "serverless",
		Version:               "0.0.1",
		IsDefaultCRApplicable: false,
		CustomResources:       []unstructured.Unstructured{},
	}

	clierr := Install(ctx, &client, data)
	require.Nil(t, clierr)
}

func getTestHttpServerWithResponse(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
		w.WriteHeader(http.StatusOK)
	}))
}
