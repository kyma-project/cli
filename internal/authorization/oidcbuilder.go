package authorization

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type OIDCBuilder struct {
	repository string
	clientID   string
	issuerURL  string
	name       string
	prefix     string
}

func NewOIDCBuilder(clientID, issuerURL string) *OIDCBuilder {
	return &OIDCBuilder{
		clientID:  clientID,
		issuerURL: issuerURL,
	}
}

func (b *OIDCBuilder) ForRepository(repository string) *OIDCBuilder {
	b.repository = repository
	return b
}

func (b *OIDCBuilder) ForClientID(clientID string) *OIDCBuilder {
	b.clientID = clientID
	return b
}

func (b *OIDCBuilder) ForIssuerURL(issuerURL string) *OIDCBuilder {
	b.issuerURL = issuerURL
	return b
}

func (b *OIDCBuilder) ForPrefix(prefix string) *OIDCBuilder {
	b.prefix = prefix
	return b
}

func (b *OIDCBuilder) ForName(name string) *OIDCBuilder {
	b.name = name
	return b
}

func (b *OIDCBuilder) Build() (*unstructured.Unstructured, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}

	oidc := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "authentication.gardener.cloud/v1alpha1",
			"kind":       "OpenIDConnect",
			"metadata": map[string]any{
				"name": b.getOpenIDConnectResourceName(),
			},
			"spec": map[string]any{
				"issuerURL":     b.issuerURL,
				"clientID":      b.clientID,
				"usernameClaim": "repository",
				"requiredClaims": map[string]any{
					"repository": b.repository,
				},
			},
		},
	}

	// Add usernamePrefix if provided
	if b.prefix != "" {
		spec := oidc.Object["spec"].(map[string]any)
		spec["usernamePrefix"] = b.prefix
	}

	return oidc, nil
}

func (b *OIDCBuilder) getOpenIDConnectResourceName() string {
	if b.name != "" {
		return b.name
	}

	return fmt.Sprintf("%s-%s", b.clientID, "oidc")
}

func (b *OIDCBuilder) validate() error {
	if b.repository == "" {
		return fmt.Errorf("repository can't be blank")
	}
	if !repositoryFormatValid(b.repository) {
		return fmt.Errorf("repository must be in owner/name format (e.g., kyma-project/cli)")
	}
	if b.clientID == "" {
		return fmt.Errorf("clientID can't be blank")
	}
	if b.issuerURL == "" {
		return fmt.Errorf("issuerURL can't be blank")
	}

	return nil
}

func repositoryFormatValid(repository string) bool {
	repoNameParts := strings.Split(repository, "/")
	return len(repoNameParts) == 2
}
