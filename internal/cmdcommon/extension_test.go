package cmdcommon

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func TestListFromCluster(t *testing.T) {
	t.Run("list extensions from cluster", func(t *testing.T) {
		client := k8s_fake.NewSimpleClientset(
			fixTestExtensionConfigmap("test-1"),
			fixTestExtensionConfigmap("test-2"),
			fixTestExtensionConfigmap("test-3"),
		)

		want := ExtensionList{
			fixTestExtension("test-1"),
			fixTestExtension("test-2"),
			fixTestExtension("test-3"),
		}

		got, err := ListExtensions(context.Background(), client)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("missing rootCommand error", func(t *testing.T) {
		client := k8s_fake.NewSimpleClientset(
			&corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name: "bad-data",
					Labels: map[string]string{
						ExtensionLabelKey: ExtensionResourceLabelValue,
					},
				},
				Data: map[string]string{},
			},
		)

		got, err := ListExtensions(context.Background(), client)
		require.ErrorContains(t, err, "failed to parse configmap '/bad-data': missing .data.rootCommand field")
		require.Nil(t, got)
	})

	t.Run("skip optional fields", func(t *testing.T) {
		client := k8s_fake.NewSimpleClientset(
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
		)

		want := ExtensionList{
			{
				RootCommand: RootCommand{
					Name:            "test-command",
					Description:     "test-description",
					DescriptionLong: "test-description-long",
				},
			},
		}

		got, err := ListExtensions(context.Background(), client)
		require.NoError(t, err)
		require.Equal(t, want, got)
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
singular: testkind
plural: testkinds
`,
			ExtensionGenericCommandsKey: `
explain:
  description: test-description
  descriptionLong: test-description-long
  output: test-explain-output
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
		Resource: &ResourceInfo{
			Scope:    NamespacedScope,
			Kind:     "TestKind",
			Group:    "test.group",
			Version:  "v1",
			Singular: "testkind",
			Plural:   "testkinds",
		},
		TemplateCommands: &TemplateCommands{
			ExplainCommand: &ExplainCommand{
				Description:     "test-description",
				DescriptionLong: "test-description-long",
				Output:          "test-explain-output",
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
