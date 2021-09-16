package deploy

import (
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

const any = "DO_NOT_CHECK_OVERRIDE_VALUE"

func TestMergeOverrides(t *testing.T) {
	testCases := []struct {
		summary     string
		values      []string
		expected    map[string]interface{}
		expectedErr bool
	}{
		{
			summary: "single value",
			values:  []string{"component.key=foo"},
			expected: map[string]interface{}{
				"global.domainName":         any,
				"global.tlsCrt":             any,
				"global.tlsKey":             any,
				"global.ingress.domainName": any,
				"component.key":             "foo",
			},
		},
		{
			summary: "single value comma separated",
			values:  []string{"component.key=foo,component.inner.key=bar"},
			expected: map[string]interface{}{
				"global.domainName":         any,
				"global.tlsCrt":             any,
				"global.tlsKey":             any,
				"global.ingress.domainName": any,
				"component.key":             "foo",
				"component.inner.key":       "bar",
			},
		},
		{
			summary: "multiple values",
			values:  []string{"component.key=foo", "component.inner.key=bar"},
			expected: map[string]interface{}{
				"global.domainName":         any,
				"global.tlsCrt":             any,
				"global.tlsKey":             any,
				"global.ingress.domainName": any,
				"component.key":             "foo",
				"component.inner.key":       "bar",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.summary, func(t *testing.T) {
			opts := &Options{
				Values: tc.values,
			}
			ovs, err := mergeOverrides(opts, &workspace.Workspace{
				InstallationResourceDir: "testdata",
			}, fake.NewSimpleClientset())

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				requireEqualOverrides(t, tc.expected, ovs)
			}
		})
	}
}

func requireEqualOverrides(t *testing.T, expected, actual map[string]interface{}) {
	var actualKeys, expectedKeys []string
	for k := range actual {
		actualKeys = append(actualKeys, k)
	}
	for k := range expected {
		expectedKeys = append(expectedKeys, k)
	}

	require.ElementsMatchf(t, expectedKeys, actualKeys, "key mismatch")
	for key, expected := range expected {
		value, exists := actual[key]
		require.Truef(t, exists, "not found %s", key)
		if expected != any {
			require.Equal(t, expected, value)
		}
	}
}
