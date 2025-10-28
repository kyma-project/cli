package authorization_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/authorization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOIDCBuilder(t *testing.T) {
	builder := authorization.NewOIDCBuilder("client-id", "issuer-url")

	assert.NotNil(t, builder)
}

func TestOIDCBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *authorization.OIDCBuilder
		expected map[string]any
	}{
		{
			name: "minimal valid configuration",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "https://token.actions.githubusercontent.com").
					ForRepository("kyma-project/cli")
			},
			expected: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata": map[string]any{
					"name": "test-client",
				},
				"spec": map[string]any{
					"issuerURL":     "https://token.actions.githubusercontent.com",
					"clientID":      "test-client",
					"usernameClaim": "repository",
					"requiredClaims": map[string]any{
						"repository": "kyma-project/cli",
					},
					"usernamePrefix": "test-client/",
				},
			},
		},
		{
			name: "with custom name",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("client-123", "https://issuer.example.com").
					ForRepository("owner/repo").
					ForName("custom-oidc-name")
			},
			expected: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata": map[string]any{
					"name": "custom-oidc-name",
				},
				"spec": map[string]any{
					"issuerURL":     "https://issuer.example.com",
					"clientID":      "client-123",
					"usernameClaim": "repository",
					"requiredClaims": map[string]any{
						"repository": "owner/repo",
					},
					"usernamePrefix": "custom-oidc-name/",
				},
			},
		},
		{
			name: "with username prefix",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("prefix-client", "https://auth.example.com").
					ForRepository("test/project").
					ForPrefix("gh-oidc:")
			},
			expected: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata": map[string]any{
					"name": "prefix-client",
				},
				"spec": map[string]any{
					"issuerURL":      "https://auth.example.com",
					"clientID":       "prefix-client",
					"usernameClaim":  "repository",
					"usernamePrefix": "gh-oidc:",
					"requiredClaims": map[string]any{
						"repository": "test/project",
					},
				},
			},
		},
		{
			name: "full configuration",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("full-client-id", "https://full.issuer.com").
					ForRepository("full/example").
					ForName("full-custom-name").
					ForPrefix("full-prefix:")
			},
			expected: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata": map[string]any{
					"name": "full-custom-name",
				},
				"spec": map[string]any{
					"issuerURL":      "https://full.issuer.com",
					"clientID":       "full-client-id",
					"usernameClaim":  "repository",
					"usernamePrefix": "full-prefix:",
					"requiredClaims": map[string]any{
						"repository": "full/example",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()

			result, err := builder.Build()

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result.Object)
		})
	}
}

func TestOIDCBuilder_Build_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *authorization.OIDCBuilder
		expectedError string
	}{
		{
			name: "missing repository",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "https://example.com")
			},
			expectedError: "repository can't be blank",
		},
		{
			name: "missing clientID",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("", "https://example.com").
					ForRepository("owner/repo")
			},
			expectedError: "clientID can't be blank",
		},
		{
			name: "missing issuerURL",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "").
					ForRepository("owner/repo")
			},
			expectedError: "issuerURL can't be blank",
		},
		{
			name: "invalid repository format - no slash",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "https://example.com").
					ForRepository("invalid-repo-name")
			},
			expectedError: "repository must be in owner/name format (e.g., kyma-project/cli)",
		},
		{
			name: "invalid repository format - multiple slashes",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "https://example.com").
					ForRepository("owner/repo/extra")
			},
			expectedError: "repository must be in owner/name format (e.g., kyma-project/cli)",
		},
		{
			name: "invalid repository format - empty string",
			setup: func() *authorization.OIDCBuilder {
				return authorization.NewOIDCBuilder("test-client", "https://example.com").
					ForRepository("")
			},
			expectedError: "repository can't be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()

			result, err := builder.Build()

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
