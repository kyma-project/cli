package cmdcommon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func TestListFromCluster(t *testing.T) {
	t.Run("list extensions from cluster", func(t *testing.T) {
		kubeClientConfig := &KubeClientConfig{
			KubeClient: &fake.KubeClient{
				TestKubernetesInterface: k8s_fake.NewSimpleClientset(
					fixTestExtensionConfigmap("test-1"),
					fixTestExtensionConfigmap("test-2"),
					fixTestExtensionConfigmap("test-3"),
				),
			},
		}

		want := ExtensionList{
			fixTestExtension("test-1"),
			fixTestExtension("test-2"),
			fixTestExtension("test-3"),
		}

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		got := newExtensionsConfig(kymaConfig)
		require.Equal(t, want, got.extensions)
		require.Empty(t, got.parseErrors)
	})

	t.Run("extensions duplications warning", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = append(os.Args, "--show-extensions-error")
		defer func() { os.Args = oldArgs }()

		kubeClientConfig := &KubeClientConfig{
			KubeClient: &fake.KubeClient{
				TestKubernetesInterface: k8s_fake.NewSimpleClientset(
					fixTestExtensionConfigmapWithCMName("test-1", "test-1"),
					fixTestExtensionConfigmapWithCMName("test-2", "test-1"),
					fixTestExtensionConfigmapWithCMName("test-3", "test-1"),
				),
			},
		}

		want := ExtensionList{
			fixTestExtension("test-1"),
		}

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		wantWarning :=
			"failed to validate configmap '/test-2': extension with rootCommand.name='test-1' already exists\n" +
				"failed to validate configmap '/test-3': extension with rootCommand.name='test-1' already exists"

		got := newExtensionsConfig(kymaConfig)

		require.Equal(t, want, got.extensions)
		require.Equal(t, wantWarning, got.parseErrors.Error())
	})

	t.Run("missing rootCommand error", func(t *testing.T) {
		kubeClientConfig := &KubeClientConfig{
			KubeClient: &fake.KubeClient{
				TestKubernetesInterface: k8s_fake.NewSimpleClientset(
					&corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Name: "bad-data",
							Labels: map[string]string{
								ExtensionLabelKey: ExtensionResourceLabelValue,
							},
						},
						Data: map[string]string{},
					},
				),
			},
		}

		wantWarning :=
			"failed to parse configmap '/bad-data': missing .data.rootCommand field"

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		got := newExtensionsConfig(kymaConfig)
		require.Equal(t, wantWarning, got.parseErrors.Error())
		require.Empty(t, got.extensions)
	})

	t.Run("skip optional fields", func(t *testing.T) {
		kubeClientConfig := &KubeClientConfig{
			KubeClient: &fake.KubeClient{
				TestKubernetesInterface: k8s_fake.NewSimpleClientset(
					&corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Name: "bad-data",
							Labels: map[string]string{
								ExtensionLabelKey: ExtensionResourceLabelValue,
							},
						},
						Data: map[string]string{
							ExtensionRootCommandKey: `
name: test-command
description: test-description
descriptionLong: test-description-long
`,
						},
					},
				),
			},
		}

		want := ExtensionList{
			{
				RootCommand: types.RootCommand{
					Name:            "test-command",
					Description:     "test-description",
					DescriptionLong: "test-description-long",
				},
			},
		}

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		got := newExtensionsConfig(kymaConfig)
		require.Empty(t, got.parseErrors)
		require.Equal(t, want, got.extensions)
	})

	t.Run("extensions warning message display", func(t *testing.T) {
		kubeClientConfig := &KubeClientConfig{
			KubeClient: &fake.KubeClient{
				TestKubernetesInterface: k8s_fake.NewSimpleClientset(
					&corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Name: "bad-data",
							Labels: map[string]string{
								ExtensionLabelKey: ExtensionResourceLabelValue,
							},
						},
						Data: map[string]string{},
					},
				),
			},
		}
		warnBuf := bytes.NewBuffer([]byte{})

		wantWarning :=
			"Extensions Warning:\nfailed to fetch all extensions from the cluster. Use the '--show-extensions-error' flag to see more details.\n\n"

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}
		kymaExtensionsConfig := newExtensionsConfig(kymaConfig)

		kymaExtensionsConfig.DisplayExtensionsErrors(warnBuf)
		require.Equal(t, wantWarning, warnBuf.String())
	})
}

func fixTestExtensionConfigmap(name string) *corev1.ConfigMap {
	return fixTestExtensionConfigmapWithCMName(name, name)
}

func fixTestExtensionConfigmapWithCMName(cmName, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: cmName,
			Labels: map[string]string{
				ExtensionLabelKey: ExtensionResourceLabelValue,
			},
		},
		Data: map[string]string{
			ExtensionRootCommandKey: fmt.Sprintf(`
name: %s
description: test-description
descriptionLong: test-description-long
`, name),
			ExtensionResourceInfoKey: `
scope: namespace
kind: TestKind
group: test.group
version: v1
`,
			ExtensionGenericCommandsKey: `
get:
  description: test-get-description
  descriptionLong: test-get-description-long
  parameters:
  - path: ".metadata.generation"
    name: "generation"
explain:
  description: test-description
  descriptionLong: test-description-long
  output: test-explain-output
delete:
  description: test-delete-description
  descriptionLong: test-delete-description-long
create:
  description: create test resource
  descriptionLong: use this command to create test resource
  customFlags:
  - type: "string"
    name: "test-flag"
    description: "test-flag description"
    shorthand: "t"
    path: ".spec.test.field"
    default: "test-default"
    required: true
  - type: "path"
    name: "test-flag-2"
    description: "test-flag-2 description"
    shorthand: "f"
    path: ".spec.test.field2"
    default: "test-default2"
    required: false
`,
			ExtensionCoreCommandsKey: `
- actionID: test-action-id-1
- actionID: test-action-id-2
`,
		},
	}
}

func fixTestExtension(name string) Extension {
	return Extension{
		RootCommand: types.RootCommand{
			Name:            name,
			Description:     "test-description",
			DescriptionLong: "test-description-long",
		},
		Resource: &types.ResourceInfo{
			Scope:   types.NamespaceScope,
			Kind:    "TestKind",
			Group:   "test.group",
			Version: "v1",
		},
		TemplateCommands: &TemplateCommands{
			GetCommand: &types.GetCommand{
				Description:     "test-get-description",
				DescriptionLong: "test-get-description-long",
				Parameters: []types.Parameter{
					{
						Path: ".metadata.generation",
						Name: "generation",
					},
				},
			},
			ExplainCommand: &types.ExplainCommand{
				Description:     "test-description",
				DescriptionLong: "test-description-long",
				Output:          "test-explain-output",
			},
			DeleteCommand: &types.DeleteCommand{
				Description:     "test-delete-description",
				DescriptionLong: "test-delete-description-long",
			},
			CreateCommand: &types.CreateCommand{
				Description:     "create test resource",
				DescriptionLong: "use this command to create test resource",
				CustomFlags: []types.CustomFlag{
					{
						Type:         types.StringCustomFlagType,
						Name:         "test-flag",
						Description:  "test-flag description",
						Shorthand:    "t",
						Path:         ".spec.test.field",
						DefaultValue: "test-default",
						Required:     true,
					},
					{
						Type:         types.PathCustomFlagType,
						Name:         "test-flag-2",
						Description:  "test-flag-2 description",
						Shorthand:    "f",
						Path:         ".spec.test.field2",
						DefaultValue: "test-default2",
						Required:     false,
					},
				},
			},
		},
		CoreCommands: []CoreCommandInfo{
			{
				ActionID: "test-action-id-1",
			},
			{
				ActionID: "test-action-id-2",
			},
		},
	}
}
