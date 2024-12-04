package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type RootlessdynamicMock struct {
	ReturnErr      error
	AppliedObjects []unstructured.Unstructured
}

func (m *RootlessdynamicMock) Apply(_ context.Context, obj *unstructured.Unstructured) error {
	m.AppliedObjects = append(m.AppliedObjects, *obj)
	return m.ReturnErr
}

func (m *RootlessdynamicMock) ApplyMany(_ context.Context, objs []unstructured.Unstructured) error {
	return m.ReturnErr
}

func (m *RootlessdynamicMock) Get(_ context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return obj, m.ReturnErr
}

func (m *RootlessdynamicMock) Remove(_ context.Context, obj *unstructured.Unstructured) error {
	return m.ReturnErr
}

func (m *RootlessdynamicMock) RemoveMany(_ context.Context, objs []unstructured.Unstructured) error {
	return m.ReturnErr
}
