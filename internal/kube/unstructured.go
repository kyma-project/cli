package kube

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ToUnstructured converts the given data to an Unstructured object
func ToUnstructured(requestData interface{}, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&requestData)
	if err != nil {
		return nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: u}
	unstructuredObj.SetGroupVersionKind(gvk)

	return unstructuredObj, nil
}

// FromUnstructured converts the given Unstructured object to the given object
func FromUnstructured(u *unstructured.Unstructured, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
}
