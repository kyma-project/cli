package modules

import (
	"bytes"
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	testKedaCR = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test/v1",
			"kind":       "Keda",
			"metadata": map[string]interface{}{
				"name":      "default",
				"namespace": "kyma-system",
			},
		},
	}
	testKedaModuleTemplate = kyma.ModuleTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"operator.kyma-project.io/managed-by": "kyma",
			},
		},
		Spec: kyma.ModuleTemplateSpec{
			ModuleName: "keda",
			Version:    "1.0.0",
		},
	}
	testKedaModuleReleaseMeta = kyma.ModuleReleaseMeta{
		Spec: kyma.ModuleReleaseMetaSpec{
			ModuleName: "keda",
			Channels: []kyma.ChannelVersionAssignment{
				{
					Channel: "fast",
					Version: "1.0.0",
				},
			},
		},
	}
)

func TestEnable(t *testing.T) {
	t.Run("enable module", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}

		expectedEnabledModule := fake.FakeEnabledModule{
			Name:                 "keda",
			Channel:              "fast",
			CustomResourcePolicy: kyma.CustomResourcePolicyCreateAndDelete,
		}

		repo := &modulesfake.ModuleTemplatesRepo{}

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", true)
		require.Nil(t, err)
		require.Equal(t, "adding the keda module to the Kyma CR\nkeda module enabled\n", buffer.String())
		require.Equal(t, []fake.FakeEnabledModule{expectedEnabledModule}, kymaClient.EnabledModules)
	})

	t.Run("enable module for default channel", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnDefaultKyma: kyma.Kyma{
				Spec: kyma.KymaSpec{
					Channel: "fast",
				},
			},
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}

		expectedEnabledModule := fake.FakeEnabledModule{
			Name:                 "keda",
			Channel:              "",
			CustomResourcePolicy: kyma.CustomResourcePolicyCreateAndDelete,
		}

		repo := &modulesfake.ModuleTemplatesRepo{}

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "", true)
		require.Nil(t, err)
		require.Equal(t, "adding the keda module to the Kyma CR\nkeda module enabled\n", buffer.String())
		require.Equal(t, []fake.FakeEnabledModule{expectedEnabledModule}, kymaClient.EnabledModules)
	})

	t.Run("enable module and add custom cr", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}
		rootlessDynamicClient := fake.RootlessDynamicClient{}
		client := fake.KubeClient{
			TestKymaInterface:            &kymaClient,
			TestRootlessDynamicInterface: &rootlessDynamicClient,
		}

		expectedEnabledModule := fake.FakeEnabledModule{
			Name:                 "keda",
			Channel:              "fast",
			CustomResourcePolicy: kyma.CustomResourcePolicyIgnore,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", false, testKedaCR)
		require.Nil(t, err)
		require.Equal(t, "adding the keda module to the Kyma CR\nwaiting for module to be ready\napplying kyma-system/default cr\nkeda module enabled\n", buffer.String())
		require.Equal(t, []fake.FakeEnabledModule{expectedEnabledModule}, kymaClient.EnabledModules)
		require.Equal(t, []unstructured.Unstructured{testKedaCR}, rootlessDynamicClient.ApplyObjs)
	})

	t.Run("failed to get module from catalog", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: errors.New("test error"),
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		hints := []string{
			"make sure you provide a valid module name and channel (or version)",
			"to list available modules, call the `kyma module catalog` command",
			"to pull available modules, call the `kyma module pull` command",
		}

		expectedCliErr := clierror.Wrap(
			errors.New("failed to list all ModuleTemplate resources from the target Kyma environment: test error"),
			clierror.New("unknown module name or channel", hints...),
		)

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", true)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("failed to get module that is not available", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{},
			},
		}

		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		hints := []string{
			"make sure you provide a valid module name and channel (or version)",
			"to list available modules, call the `kyma module catalog` command",
			"to pull available modules, call the `kyma module pull` command",
		}

		expectedCliErr := clierror.Wrap(
			errors.New("the keda module is not available in the catalog"),
			clierror.New("unknown module name or channel", hints...),
		)

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", true)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("failed to get module in given channel", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}

		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		hints := []string{
			"make sure you provide a valid module name and channel (or version)",
			"to list available modules, call the `kyma module catalog` command",
			"to pull available modules, call the `kyma module pull` command",
		}

		expectedCliErr := clierror.Wrap(
			errors.New("the keda module is not available in the regular channel"),
			clierror.New("unknown module name or channel", hints...),
		)

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "regular", true)
		require.Equal(t, expectedCliErr, err)
	})

	t.Run("failed to wait for module to be ready", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnWaitForModuleErr: errors.New("test error"),
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}
		client := fake.KubeClient{
			TestKymaInterface: &kymaClient,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to check the module state"),
		)

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", false, testKedaCR)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "adding the keda module to the Kyma CR\nwaiting for module to be ready\n", buffer.String())
	})

	t.Run("failed to apply custom resource", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		kymaClient := fake.KymaClient{
			ReturnErr: nil,
			ReturnModuleTemplateList: kyma.ModuleTemplateList{
				Items: []kyma.ModuleTemplate{testKedaModuleTemplate},
			},
			ReturnModuleReleaseMetaList: kyma.ModuleReleaseMetaList{
				Items: []kyma.ModuleReleaseMeta{testKedaModuleReleaseMeta},
			},
		}
		rootlessDynamicClient := fake.RootlessDynamicClient{
			ReturnErr: errors.New("test error"),
		}
		client := fake.KubeClient{
			TestKymaInterface:            &kymaClient,
			TestRootlessDynamicInterface: &rootlessDynamicClient,
		}
		repo := &modulesfake.ModuleTemplatesRepo{}

		expectedCliErr := clierror.Wrap(
			errors.New("test error"),
			clierror.New("failed to apply a custom CR from path"),
		)

		err := enable(out.NewToWriter(buffer), context.Background(), &client, repo, "keda", "fast", false, testKedaCR)
		require.Equal(t, expectedCliErr, err)
		require.Equal(t, "adding the keda module to the Kyma CR\nwaiting for module to be ready\napplying kyma-system/default cr\n", buffer.String())
	})
}
