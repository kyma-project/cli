package authorization

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	USER            = "User"
	GROUP           = "Group"
	SERVICE_ACCOUNT = "ServiceAccount"
)

type SubjectKind struct {
	name string
}

func NewSubjectKindFrom(name string) (*SubjectKind, error) {
	if strings.EqualFold(name, USER) {
		return &SubjectKind{USER}, nil
	}
	if strings.EqualFold(name, GROUP) {
		return &SubjectKind{GROUP}, nil
	}
	if strings.EqualFold(name, SERVICE_ACCOUNT) {
		return &SubjectKind{SERVICE_ACCOUNT}, nil
	}

	return nil, fmt.Errorf("invalid subjectKind: %s", name)
}

type RBACBuilder struct {
	prefix      string
	namespace   string
	role        string
	clusterrole string

	subjectKind *SubjectKind
	subjectName string
	bindingName string

	serviceAccountNamespace string
}

func NewRBACBuilder() *RBACBuilder {
	return &RBACBuilder{}
}

func (b *RBACBuilder) ForSubjectKind(subjectKind *SubjectKind) *RBACBuilder {
	b.subjectKind = subjectKind
	return b
}

func (b *RBACBuilder) ForSubjectName(subjectName string) *RBACBuilder {
	b.subjectName = subjectName
	return b
}

func (b *RBACBuilder) ForBindingName(bindingName string) *RBACBuilder {
	b.bindingName = bindingName
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

func (b *RBACBuilder) ForServiceAccountNamespace(saNamespace string) *RBACBuilder {
	b.serviceAccountNamespace = saNamespace
	return b
}

func (b *RBACBuilder) BuildClusterRoleBinding() (*unstructured.Unstructured, clierror.Error) {
	if err := b.validateForClusterRoleBinding(); err != nil {
		return nil, err
	}

	return b.buildClusterRoleBinding(), nil
}

func (b *RBACBuilder) BuildRoleBinding() (*unstructured.Unstructured, clierror.Error) {
	if err := b.validateForRoleBinding(); err != nil {
		return nil, err
	}

	return b.buildRoleBinding(), nil
}

func (b *RBACBuilder) validateForClusterRoleBinding() clierror.Error {
	if b.subjectKind == nil {
		return clierror.New("subjectKind is required")
	}

	if b.subjectName == "" {
		return clierror.New("subjectName is required")
	}

	if b.clusterrole == "" {
		return clierror.New("clusterrole is required for ClusterRoleBinding")
	}

	if err := b.validateBindingName(); err != nil {
		return err
	}

	if b.subjectKind != nil && b.subjectKind.name == SERVICE_ACCOUNT && b.namespace == "" && b.serviceAccountNamespace == "" {
		return clierror.New("namespace is required for service account subject", "Make use of '--sa-namespace' flag to define namespace for ServiceAccount")
	}

	return nil
}

func (b *RBACBuilder) validateForRoleBinding() clierror.Error {
	if b.subjectKind == nil {
		return clierror.New("subjectKind is required")
	}
	if b.subjectName == "" {
		return clierror.New("subjectName is required")
	}
	if b.role == "" && b.clusterrole == "" {
		return clierror.New("specify either Role or ClusterRole for RoleBinding")
	}
	if b.role != "" && b.clusterrole != "" {
		return clierror.New("cannot specify both Role and ClusterRole for RoleBinding")
	}
	if b.namespace == "" && b.clusterrole != "" {
		return clierror.New(
			fmt.Sprintf("failed to apply binding for the '%s' ClusterRole", b.clusterrole),
			"either specify a namespace or enable the cluster-wide flag for ClusterRoleBinding",
		)
	}
	if b.namespace == "" {
		return clierror.New("provide namespace for RoleBinding")
	}
	if err := b.validateBindingName(); err != nil {
		return err
	}
	if b.subjectKind != nil && b.subjectKind.name == SERVICE_ACCOUNT && b.namespace == "" && b.serviceAccountNamespace == "" {
		return clierror.New("provide namespace for ServiceAccount subject")
	}

	return nil
}

func (b *RBACBuilder) validateBindingName() clierror.Error {
	if b.bindingName == "" {
		return clierror.New("binding name is required")
	}

	// Kubernetes resource names must follow DNS-1123 subdomain rules:
	// - Must be no more than 253 characters
	// - Contain only lowercase alphanumeric characters, '-' or '.'
	// - Start with an alphanumeric character
	// - End with an alphanumeric character
	if len(b.bindingName) > 253 {
		return clierror.New(
			fmt.Sprintf("binding name '%s' is too long (%d characters)", b.bindingName, len(b.bindingName)),
			"Kubernetes resource names must not exceed 253 characters",
		)
	}

	// Check if it starts with alphanumeric
	if len(b.bindingName) > 0 {
		firstChar := b.bindingName[0]
		if (firstChar < 'a' || firstChar > 'z') && (firstChar < '0' || firstChar > '9') {
			return clierror.New(
				fmt.Sprintf("binding name '%s' must start with a lowercase alphanumeric character", b.bindingName),
			)
		}
	}

	// Check if it ends with alphanumeric
	if len(b.bindingName) > 0 {
		lastChar := b.bindingName[len(b.bindingName)-1]
		if (lastChar < 'a' || lastChar > 'z') && (lastChar < '0' || lastChar > '9') {
			return clierror.New(
				fmt.Sprintf("binding name '%s' must end with a lowercase alphanumeric character", b.bindingName),
			)
		}
	}

	// Check all characters are valid
	for i, char := range b.bindingName {
		isLowerAlpha := char >= 'a' && char <= 'z'
		isDigit := char >= '0' && char <= '9'
		isDash := char == '-'
		isDot := char == '.'

		if !isLowerAlpha && !isDigit && !isDash && !isDot {
			return clierror.New(
				fmt.Sprintf("binding name '%s' contains invalid character at position %d", b.bindingName, i),
				"Binding names must contain only lowercase alphanumeric characters, '-' or '.'",
			)
		}
	}

	return nil
}

func (b *RBACBuilder) buildClusterRoleBinding() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRoleBinding",
			"metadata": map[string]any{
				"name": b.bindingName,
			},
			"subjects": b.buildRoleBindingSubject(),
			"roleRef": map[string]any{
				"kind":     "ClusterRole",
				"name":     b.clusterrole,
				"apiGroup": "rbac.authorization.k8s.io",
			},
		},
	}
}

func (b *RBACBuilder) buildRoleBinding() *unstructured.Unstructured {
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
				"name":      b.bindingName,
				"namespace": b.namespace,
			},
			"subjects": b.buildRoleBindingSubject(),
			"roleRef": map[string]any{
				"kind":     roleKind,
				"name":     roleName,
				"apiGroup": "rbac.authorization.k8s.io",
			},
		},
	}
}

func (b *RBACBuilder) buildRoleBindingSubject() []map[string]any {
	if b.subjectKind.name == SERVICE_ACCOUNT {
		saNamespace := b.serviceAccountNamespace
		if saNamespace == "" {
			saNamespace = b.namespace
		}

		return []map[string]any{
			{
				"kind":      b.subjectKind.name,
				"name":      b.prefix + b.subjectName,
				"namespace": saNamespace,
			},
		}
	}

	return []map[string]any{
		{
			"kind": b.subjectKind.name,
			"name": b.prefix + b.subjectName,
		},
	}
}
