package authorization

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type RBACBuilder struct {
	repository  string
	prefix      string
	namespace   string
	role        string
	clusterrole string
}

func NewRBACBuilder() *RBACBuilder {
	return &RBACBuilder{}
}

func (b *RBACBuilder) ForRepository(repository string) *RBACBuilder {
	b.repository = repository
	return b
}

func (b *RBACBuilder) ForPrefix(prefix string) *RBACBuilder {
	b.prefix = prefix
	return b
}

func (b *RBACBuilder) ForNamespace(namespace string) *RBACBuilder {
	b.namespace = namespace
	return b
}

func (b *RBACBuilder) ForRole(role string) *RBACBuilder {
	b.role = role
	return b
}

func (b *RBACBuilder) ForClusterRole(clusterRole string) *RBACBuilder {
	b.clusterrole = clusterRole
	return b
}

func (b *RBACBuilder) BuildClusterRoleBinding() (*unstructured.Unstructured, clierror.Error) {
	if err := b.validateForClusterRoleBinding(); err != nil {
		return nil, err
	}

	subjectName := b.getSubjectName()
	bindingName := b.getClusterRoleBindingName()

	return b.buildClusterRoleBinding(bindingName, subjectName), nil
}

func (b *RBACBuilder) BuildRoleBinding() (*unstructured.Unstructured, clierror.Error) {
	if err := b.validateForRoleBinding(); err != nil {
		return nil, err
	}

	subjectName := b.getSubjectName()
	bindingName := b.getRoleBindingName()

	return b.buildRoleBinding(bindingName, subjectName), nil
}

func (b *RBACBuilder) validateForClusterRoleBinding() clierror.Error {
	if b.repository == "" {
		return clierror.New("repository is required")
	}

	if !b.isRepoNameValid() {
		return clierror.New("repository must be in owner/name format (e.g., kyma-project/cli)")
	}

	if b.clusterrole == "" {
		return clierror.New("clusterrole is required for ClusterRoleBinding")
	}

	return nil
}

func (b *RBACBuilder) validateForRoleBinding() clierror.Error {
	if b.repository == "" {
		return clierror.New("repository is required")
	}

	if !b.isRepoNameValid() {
		return clierror.New("repository must be in owner/name format (e.g., kyma-project/cli)")
	}

	if b.role == "" && b.clusterrole == "" {
		return clierror.New("either role or clusterrole must be specified for RoleBinding")
	}
	if b.role != "" && b.clusterrole != "" {
		return clierror.New("cannot specify both role and clusterrole for RoleBinding")
	}
	if b.namespace == "" && b.clusterrole != "" {
		return clierror.New(
			fmt.Sprintf("failed to apply binding for the '%s' ClusterRole", b.clusterrole),
			"either specify a namespace or enable cluster-wide flag for ClusterRoleBinding",
		)
	}
	if b.namespace == "" {
		return clierror.New("namespace is required for RoleBinding")
	}

	return nil
}

func (b *RBACBuilder) isRepoNameValid() bool {
	repoNameParts := strings.Split(b.repository, "/")
	return len(repoNameParts) == 2
}

func (b *RBACBuilder) getSubjectName() string {
	return b.prefix + b.repository
}

func (b *RBACBuilder) getClusterRoleBindingName() string {
	sanitizedRepo := strings.ReplaceAll(b.repository, "/", "-")
	return fmt.Sprintf("%s-%s-binding", sanitizedRepo, b.clusterrole)
}

func (b *RBACBuilder) getRoleBindingName() string {
	sanitizedRepo := strings.ReplaceAll(b.repository, "/", "-")

	var roleName string
	if b.role != "" {
		roleName = b.role
	} else {
		roleName = b.clusterrole
	}

	return fmt.Sprintf("%s-%s-binding", sanitizedRepo, roleName)
}

func (b *RBACBuilder) buildClusterRoleBinding(bindingName, subjectName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRoleBinding",
			"metadata": map[string]any{
				"name": bindingName,
			},
			"subjects": []map[string]any{
				{
					"kind": "User",
					"name": subjectName,
				},
			},
			"roleRef": map[string]any{
				"kind":     "ClusterRole",
				"name":     b.clusterrole,
				"apiGroup": "rbac.authorization.k8s.io",
			},
		},
	}
}

func (b *RBACBuilder) buildRoleBinding(bindingName, subjectName string) *unstructured.Unstructured {
	roleKind := "Role"
	roleName := b.role
	if b.clusterrole != "" {
		roleKind = "ClusterRole"
		roleName = b.clusterrole
	}

	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]any{
				"name":      bindingName,
				"namespace": b.namespace,
			},
			"subjects": []map[string]any{
				{
					"kind": "User",
					"name": subjectName,
				},
			},
			"roleRef": map[string]any{
				"kind":     roleKind,
				"name":     roleName,
				"apiGroup": "rbac.authorization.k8s.io",
			},
		},
	}
}
