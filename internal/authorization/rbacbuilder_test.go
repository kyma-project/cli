package authorization_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewRBACBuilder(t *testing.T) {
	builder := authorization.NewRBACBuilder()
	assert.NotNil(t, builder)
	assert.IsType(t, &authorization.RBACBuilder{}, builder)
}

func TestRBACBuilder_BuildClusterRoleBinding(t *testing.T) {
	tests := []struct {
		name           string
		setupBuilder   func() *authorization.RBACBuilder
		expectedError  string
		validateResult func(t *testing.T, result *unstructured.Unstructured)
	}{
		{
			name: "successful build with minimal configuration",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("kyma-project/cli").
					ForClusterRole("cluster-admin")
			},
			validateResult: func(t *testing.T, result *unstructured.Unstructured) {
				assert.Equal(t, "rbac.authorization.k8s.io/v1", result.GetAPIVersion())
				assert.Equal(t, "ClusterRoleBinding", result.GetKind())
				assert.Equal(t, "kyma-project-cli-cluster-admin-binding", result.GetName())

				subjects, found, err := unstructured.NestedFieldNoCopy(result.Object, "subjects")
				require.NoError(t, err)
				require.True(t, found)
				subjectsSlice := subjects.([]map[string]any)
				require.Len(t, subjectsSlice, 1)

				subject := subjectsSlice[0]
				assert.Equal(t, "User", subject["kind"])
				assert.Equal(t, "kyma-project/cli", subject["name"])

				roleRef, found, err := unstructured.NestedMap(result.Object, "roleRef")
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, "ClusterRole", roleRef["kind"])
				assert.Equal(t, "cluster-admin", roleRef["name"])
				assert.Equal(t, "rbac.authorization.k8s.io", roleRef["apiGroup"])
			},
		},
		{
			name: "successful build with prefix",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("owner/repo").
					ForClusterRole("reader").
					ForPrefix("github:")
			},
			validateResult: func(t *testing.T, result *unstructured.Unstructured) {
				assert.Equal(t, "owner-repo-reader-binding", result.GetName())

				subjects, found, err := unstructured.NestedFieldNoCopy(result.Object, "subjects")
				require.NoError(t, err)
				require.True(t, found)
				subjectsSlice := subjects.([]map[string]any)
				require.Len(t, subjectsSlice, 1)

				subject := subjectsSlice[0]
				assert.Equal(t, "github:owner/repo", subject["name"])
			},
		},
		{
			name: "error when repository is missing",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().ForClusterRole("cluster-admin")
			},
			expectedError: "repository is required",
		},
		{
			name: "error when repository format is invalid",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("invalid-repo-format").
					ForClusterRole("cluster-admin")
			},
			expectedError: "repository must be in owner/name format (e.g., kyma-project/cli)",
		},
		{
			name: "error when clusterrole is missing",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().ForRepository("owner/repo")
			},
			expectedError: "clusterrole is required for ClusterRoleBinding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setupBuilder()
			result, err := builder.BuildClusterRoleBinding()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				tt.validateResult(t, result)
			}
		})
	}
}

func TestRBACBuilder_BuildRoleBinding(t *testing.T) {
	tests := []struct {
		name           string
		setupBuilder   func() *authorization.RBACBuilder
		expectedError  string
		validateResult func(t *testing.T, result *unstructured.Unstructured)
	}{
		{
			name: "successful build with Role",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("kyma-project/cli").
					ForRole("reader").
					ForNamespace("default")
			},
			validateResult: func(t *testing.T, result *unstructured.Unstructured) {
				assert.Equal(t, "rbac.authorization.k8s.io/v1", result.GetAPIVersion())
				assert.Equal(t, "RoleBinding", result.GetKind())
				assert.Equal(t, "kyma-project-cli-reader-binding", result.GetName())
				assert.Equal(t, "default", result.GetNamespace())

				subjects, found, err := unstructured.NestedFieldNoCopy(result.Object, "subjects")
				require.NoError(t, err)
				require.True(t, found)
				subjectsSlice := subjects.([]map[string]any)
				require.Len(t, subjectsSlice, 1)

				subject := subjectsSlice[0]
				assert.Equal(t, "User", subject["kind"])
				assert.Equal(t, "kyma-project/cli", subject["name"])

				roleRef, found, err := unstructured.NestedMap(result.Object, "roleRef")
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, "Role", roleRef["kind"])
				assert.Equal(t, "reader", roleRef["name"])
				assert.Equal(t, "rbac.authorization.k8s.io", roleRef["apiGroup"])
			},
		},
		{
			name: "successful build with ClusterRole",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("owner/repo").
					ForClusterRole("cluster-admin").
					ForNamespace("kube-system")
			},
			validateResult: func(t *testing.T, result *unstructured.Unstructured) {
				assert.Equal(t, "owner-repo-cluster-admin-binding", result.GetName())
				assert.Equal(t, "kube-system", result.GetNamespace())

				roleRef, found, err := unstructured.NestedMap(result.Object, "roleRef")
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, "ClusterRole", roleRef["kind"])
				assert.Equal(t, "cluster-admin", roleRef["name"])
			},
		},
		{
			name: "successful build with prefix",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("test/repo").
					ForRole("editor").
					ForNamespace("test-ns").
					ForPrefix("system:")
			},
			validateResult: func(t *testing.T, result *unstructured.Unstructured) {
				subjects, found, err := unstructured.NestedFieldNoCopy(result.Object, "subjects")
				require.NoError(t, err)
				require.True(t, found)
				subjectsSlice := subjects.([]map[string]any)
				require.Len(t, subjectsSlice, 1)

				subject := subjectsSlice[0]
				assert.Equal(t, "system:test/repo", subject["name"])
			},
		},
		{
			name: "error when repository is missing",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRole("reader").
					ForNamespace("default")
			},
			expectedError: "repository is required",
		},
		{
			name: "error when repository format is invalid",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("invalid").
					ForRole("reader").
					ForNamespace("default")
			},
			expectedError: "repository must be in owner/name format (e.g., kyma-project/cli)",
		},
		{
			name: "error when neither role nor clusterrole is specified",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("owner/repo").
					ForNamespace("default")
			},
			expectedError: "either role or clusterrole must be specified for RoleBinding",
		},
		{
			name: "error when both role and clusterrole are specified",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("owner/repo").
					ForRole("reader").
					ForClusterRole("cluster-admin").
					ForNamespace("default")
			},
			expectedError: "cannot specify both role and clusterrole for RoleBinding",
		},
		{
			name: "error when namespace is missing",
			setupBuilder: func() *authorization.RBACBuilder {
				return authorization.NewRBACBuilder().
					ForRepository("owner/repo").
					ForRole("reader")
			},
			expectedError: "namespace is required for RoleBinding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setupBuilder()
			result, err := builder.BuildRoleBinding()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				tt.validateResult(t, result)
			}
		})
	}
}
