package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type RootlessDynamicClient struct {
	// outputs
	ReturnErr       error
	ReturnGetErr    error
	ReturnRemoveErr error
	ReturnWatchErr  error
	ReturnGetObj    unstructured.Unstructured
	ReturnListObjs  *unstructured.UnstructuredList
	ReturnWatcher   watch.Interface

	// inputs summary
	GetObjs     []unstructured.Unstructured
	ListObjs    []unstructured.Unstructured
	RemovedObjs []unstructured.Unstructured
	ApplyObjs   []unstructured.Unstructured
}

func (m *RootlessDynamicClient) Apply(_ context.Context, obj *unstructured.Unstructured) error {
	m.ApplyObjs = append(m.ApplyObjs, *obj)
	return m.ReturnErr
}

func (m *RootlessDynamicClient) ApplyMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.ApplyObjs = append(m.ApplyObjs, objs...)
	return m.ReturnErr
}

func (m *RootlessDynamicClient) Get(_ context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	m.GetObjs = append(m.GetObjs, *obj)
	return &m.ReturnGetObj, m.ReturnGetErr
}

func (m *RootlessDynamicClient) List(_ context.Context, obj *unstructured.Unstructured) (*unstructured.UnstructuredList, error) {
	m.ListObjs = append(m.ListObjs, *obj)
	return m.ReturnListObjs, m.ReturnErr
}

func (m *RootlessDynamicClient) Remove(_ context.Context, obj *unstructured.Unstructured) error {
	m.RemovedObjs = append(m.RemovedObjs, *obj)
	return m.ReturnRemoveErr
}

func (m *RootlessDynamicClient) RemoveMany(_ context.Context, objs []unstructured.Unstructured) error {
	m.RemovedObjs = append(m.RemovedObjs, objs...)
	return m.ReturnRemoveErr
}

func (m *RootlessDynamicClient) WatchSingleResource(_ context.Context, obj *unstructured.Unstructured) (watch.Interface, error) {
	return m.ReturnWatcher, m.ReturnWatchErr
}
