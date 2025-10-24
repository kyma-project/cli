package extensions

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

const (
	testExtensionString = `
with:
  resource:
    apiVersion: v1
    kind: ConfigMap
metadata:
  name: resource
  description: manage resources
  descriptionLong: use to manage resources
subCommands:
- metadata:
    name: create
  uses: action-1
- metadata:
    name: delete
  uses: action-2
`
)

var (
	testExtension = types.Extension{
		Config: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
		},
		Metadata: types.Metadata{
			Name:            "resource",
			Description:     "manage resources",
			DescriptionLong: "use to manage resources",
		},
		SubCommands: []types.Extension{
			{
				Metadata: types.Metadata{
					Name: "create",
				},
				Action: "action-1",
			},
			{
				Metadata: types.Metadata{
					Name: "delete",
				},
				Action: "action-2",
			},
		},
	}

	testActionsMap = types.ActionsMap{
		"action-1": &mockAction{},
		"action-2": &mockAction{},
	}
)

func Test_DisplayWarnings(t *testing.T) {
	t.Run("skip on no errors", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		b := Builder{
			printer:          out.NewToWriter(buffer),
			extensionsErrors: []error{},
		}

		b.DisplayWarnings()

		require.Equal(t, "", buffer.String())
	})

	t.Run("skip on excluded command", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = append(os.Args, "help")
		defer func() { os.Args = oldArgs }()

		buffer := bytes.NewBuffer([]byte{})
		b := Builder{
			printer: out.NewToWriter(buffer),
			extensionsErrors: []error{
				errors.New("test error"),
			},
		}

		b.DisplayWarnings()

		require.Equal(t, "", buffer.String())
	})

	t.Run("display general warning", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})
		b := Builder{
			printer: out.NewToWriter(buffer),
			extensionsErrors: []error{
				errors.New("test error"),
			},
		}

		b.DisplayWarnings()

		require.Equal(t, "Extensions Warning:\n"+
			"failed to fetch all extensions from the target Kyma environment. Use the '--show-extensions-error' flag to see more details.\n\n", buffer.String())
	})

	t.Run("display warning details", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = append(os.Args, "--show-extensions-error", "true")
		defer func() { os.Args = oldArgs }()

		buffer := bytes.NewBuffer([]byte{})
		b := Builder{
			printer: out.NewToWriter(buffer),
			extensionsErrors: []error{
				errors.New("test error"),
			},
		}

		b.DisplayWarnings()

		require.Equal(t, "Extensions Warning:\ntest error\n\n", buffer.String())
	})
}

func Test_NewBuilder(t *testing.T) {
	t.Run("skip extensions", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = append(os.Args, "--skip-extensions")
		defer func() { os.Args = oldArgs }()

		b := NewBuilder(&cmdcommon.KymaConfig{})
		require.Equal(t, Builder{printer: out.Default}, *b)
	})

	t.Run("handle client error", func(t *testing.T) {
		b := NewBuilder(&cmdcommon.KymaConfig{
			Ctx: context.Background(),
			KubeClientConfig: &fakeKubeClientConfig{
				err: errors.New("client error"),
			},
		})

		require.Equal(t, []error{errors.New("client error")}, b.extensionsErrors)
	})

	t.Run("list extensions from cluster", func(t *testing.T) {
		b := NewBuilder(&cmdcommon.KymaConfig{
			Ctx: context.Background(),
			KubeClientConfig: &fakeKubeClientConfig{
				kubeClient: &kubefake.KubeClient{
					TestKubernetesInterface: k8sfake.NewClientset(
						fixTestExtensionConfigMap("cm1", testExtensionString),
					),
				},
			},
		})

		expectedExtensions := []types.ConfigmapCommandExtension{
			{
				ConfigMapName:      "cm1",
				ConfigMapNamespace: "kyma-system",
				Extension:          testExtension,
			},
		}

		require.Empty(t, b.extensionsErrors)
		require.Equal(t, expectedExtensions, b.extensions)
	})

	t.Run("handle extension duplicates", func(t *testing.T) {
		b := NewBuilder(&cmdcommon.KymaConfig{
			Ctx: context.Background(),
			KubeClientConfig: &fakeKubeClientConfig{
				kubeClient: &kubefake.KubeClient{
					TestKubernetesInterface: k8sfake.NewClientset(
						fixTestExtensionConfigMap("cm1", testExtensionString),
						fixTestExtensionConfigMap("cm2", testExtensionString),
					),
				},
			},
		})

		expectedExtensions := []types.ConfigmapCommandExtension{
			{
				ConfigMapName:      "cm1",
				ConfigMapNamespace: "kyma-system",
				Extension:          testExtension,
			},
		}

		require.Equal(t, []error{errors.NewList(errors.New("failed to validate configmap 'kyma-system/cm2': extension with name 'resource' already exists"))}, b.extensionsErrors)
		require.Equal(t, expectedExtensions, b.extensions)
	})

	t.Run("handle missing required cm value error", func(t *testing.T) {
		b := NewBuilder(&cmdcommon.KymaConfig{
			Ctx: context.Background(),
			KubeClientConfig: &fakeKubeClientConfig{
				kubeClient: &kubefake.KubeClient{
					TestKubernetesInterface: k8sfake.NewClientset(
						fixTestConfigMap("cm1", nil),
					),
				},
			},
		})

		expectedExtensions := []types.ConfigmapCommandExtension{}

		require.Equal(t, []error{errors.NewList(errors.New("failed to parse configmap 'kyma-system/cm1': missing .data.kyma-commands.yaml field"))}, b.extensionsErrors)
		require.Equal(t, expectedExtensions, b.extensions)
	})
}

func Test_Build(t *testing.T) {
	t.Run("build extension", func(t *testing.T) {
		cmd := &cobra.Command{}
		b := Builder{
			extensions: []types.ConfigmapCommandExtension{
				{
					ConfigMapName:      "cm1",
					ConfigMapNamespace: "ns",
					Extension:          testExtension,
				},
			},
		}

		b.Build(cmd, testActionsMap)

		require.Empty(t, b.extensionsErrors)
		require.Len(t, cmd.Commands(), 1)
		require.Equal(t, "resource", cmd.Commands()[0].Name())
	})

	t.Run("handle validation error", func(t *testing.T) {
		cmd := &cobra.Command{}
		b := Builder{
			extensions: []types.ConfigmapCommandExtension{
				{
					ConfigMapName:      "cm1",
					ConfigMapNamespace: "ns",
					Extension:          testExtension,
				},
				{
					ConfigMapName:      "cm2",
					ConfigMapNamespace: "ns",
					Extension: types.Extension{
						Metadata: types.Metadata{
							// missing name
						},
					},
				},
			},
		}

		b.Build(cmd, testActionsMap)

		require.Equal(t, []error{errors.New("failed to validate extension from configmap 'ns/cm2':\n  wrong .metadata: empty name")}, b.extensionsErrors)
		require.Len(t, cmd.Commands(), 1)
		require.Equal(t, "resource", cmd.Commands()[0].Name())
	})

	t.Run("handle duplicate error", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.AddCommand(&cobra.Command{
			Use: "duplicate",
		})

		b := Builder{
			extensions: []types.ConfigmapCommandExtension{
				{
					ConfigMapName:      "cm1",
					ConfigMapNamespace: "ns",
					Extension:          testExtension,
				},
				{
					ConfigMapName:      "cm2",
					ConfigMapNamespace: "ns",
					Extension: types.Extension{
						Metadata: types.Metadata{
							Name: "duplicate",
						},
					},
				},
			},
		}

		b.Build(cmd, testActionsMap)

		require.Equal(t, []error{errors.New("failed to add extension from configmap 'ns/cm2': base command with name 'duplicate' already exists")}, b.extensionsErrors)
		require.Len(t, cmd.Commands(), 2)
		require.Equal(t, "resource", cmd.Commands()[1].Name())
	})

	t.Run("handle build command error", func(t *testing.T) {
		cmd := &cobra.Command{}
		b := Builder{
			extensions: []types.ConfigmapCommandExtension{
				{
					ConfigMapName:      "cm1",
					ConfigMapNamespace: "ns",
					Extension: types.Extension{
						Metadata: types.Metadata{
							Name: "create",
						},
						Action: "action-1",
						Flags: []types.Flag{
							{
								Name:         "test-flag",
								Type:         parameters.IntCustomType,
								DefaultValue: toPtr("WRONG VALUE"),
							},
						},
					},
				},
			},
		}

		b.Build(cmd, testActionsMap)

		require.Equal(t, []error{errors.New("failed to build extension from configmap 'ns/cm1':\n" +
			"  failed to build command 'create':\n" +
			"    flag 'test-flag' error: strconv.ParseInt: parsing \"WRONG VALUE\": invalid syntax")}, b.extensionsErrors)
		require.Empty(t, cmd.Commands())
	})
}

func fixTestExtensionConfigMap(name, data string) *corev1.ConfigMap {
	return fixTestConfigMap(name, map[string]string{
		types.ExtensionCMDataKey: data,
	})
}

func fixTestConfigMap(name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kyma-system",
			Labels: map[string]string{
				types.ExtensionCMLabelKey: types.ExtensionCMLabelValue,
			},
		},
		Data: data,
	}
}

type fakeKubeClientConfig struct {
	kubeClient kube.Client
	err        error
}

func (f *fakeKubeClientConfig) GetKubeClient() (kube.Client, error) {
	return f.kubeClient, f.err
}

func (f *fakeKubeClientConfig) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	// not implemented
	return nil, nil
}

func toPtr[T any](v T) *T {
	return &v
}
