package rootlessdynamic

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Fake struct {
	// outputs
	ReturnErr      error
	ReturnGetObj   unstructured.Unstructured
	ReturnListObjs *unstructured.UnstructuredList

	// inputs summary
	GetObjs     []unstructured.Unstructured
	ListObjs    []unstructured.Unstructured
	RemovedObjs []unstructured.Unstructured
	ApplyObjs   []unstructured.Unstructured
}

func (m *Fake) Apply(_ context.Context, obj *unstructured.Unstructured) error {
	m.ApplyObjs = append(m.ApplyObjs, *obj)
	return m.ReturnErr
}

func (m *Fake) ApplyMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.ApplyObjs = append(m.ApplyObjs, objs...)
	return m.ReturnErr
}

func (m *Fake) Get(_ context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	m.GetObjs = append(m.GetObjs, *obj)
	return &m.ReturnGetObj, m.ReturnErr
}

func (m *Fake) List(_ context.Context, obj *unstructured.Unstructured) (*unstructured.UnstructuredList, error) {
	m.ListObjs = append(m.ListObjs, *obj)
	return m.ReturnListObjs, m.ReturnErr
}

func (m *Fake) Remove(_ context.Context, obj *unstructured.Unstructured) error {
	m.RemovedObjs = append(m.RemovedObjs, *obj)
	return m.ReturnErr
}

func (m *Fake) RemoveMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.RemovedObjs = append(m.RemovedObjs, objs...)
	return m.ReturnErr
}
