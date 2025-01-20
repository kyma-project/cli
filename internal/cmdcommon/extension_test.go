package cmdcommon

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/spf13/cobra"
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

		cmd := &cobra.Command{}

		want := ExtensionList{
			fixTestExtension("test-1"),
			fixTestExtension("test-2"),
			fixTestExtension("test-3"),
		}

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		got := newExtensionsConfig(kymaConfig, cmd)
		require.Equal(t, want, got.extensions)
		require.True(t, cmd.PersistentFlags().HasFlags())
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

		kymaConfig := &KymaConfig{
			Ctx:              context.Background(),
			KubeClientConfig: kubeClientConfig,
		}

		got := newExtensionsConfig(kymaConfig, &cobra.Command{})
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
				RootCommand: RootCommand{
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

		got := newExtensionsConfig(kymaConfig, &cobra.Command{})
		require.Equal(t, want, got.extensions)
	})
}

func fixTestExtensionConfigmap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
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
explain:
  description: test-description
  descriptionLong: test-description-long
  output: test-explain-output
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
		RootCommand: RootCommand{
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
			ExplainCommand: &types.ExplainCommand{
				Description:     "test-description",
				DescriptionLong: "test-description-long",
				Output:          "test-explain-output",
			},
			CreateCommand: &types.CreateCommand{
				Description:     "create test resource",
				DescriptionLong: "use this command to create test resource",
				CustomFlags: []types.CreateCustomFlag{
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
