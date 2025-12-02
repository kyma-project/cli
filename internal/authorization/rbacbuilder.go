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
			"subjects": []map[string]any{
				{
					"kind": b.subjectKind.name,
					"name": b.prefix + b.subjectName,
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
			"subjects": []map[string]any{
				{
					"kind": b.subjectKind.name,
					"name": b.prefix + b.subjectName,
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
