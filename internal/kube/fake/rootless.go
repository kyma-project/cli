package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type RootlessDynamicMock struct {
	ReturnErr      error
	AppliedObjects []unstructured.Unstructured
}

func (m *RootlessDynamicMock) Apply(_ context.Context, obj *unstructured.Unstructured) error {
	m.AppliedObjects = append(m.AppliedObjects, *obj)
	return m.ReturnErr
}

func (m *RootlessDynamicMock) ApplyMany(_ context.Context, objs []unstructured.Unstructured) error {
	return m.ReturnErr
}

func (m *RootlessDynamicMock) Get(_ context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return obj, m.ReturnErr
}

func (m *RootlessDynamicMock) Remove(_ context.Context, obj *unstructured.Unstructured) error {
	return m.ReturnErr
}

func (m *RootlessDynamicMock) RemoveMany(_ context.Context, objs []unstructured.Unstructured) error {
	return m.ReturnErr
}
