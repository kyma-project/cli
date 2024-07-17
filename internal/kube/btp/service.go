package btp

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

type Interface interface {
	GetServiceInstance(context.Context, string, string) (*ServiceInstance, error)
	GetServiceBinding(context.Context, string, string) (*ServiceBinding, error)
	CreateServiceInstance(context.Context, *ServiceInstance) error
	CreateServiceBinding(context.Context, *ServiceBinding) error
	IsBindingReady(context.Context, string, string) wait.ConditionWithContextFunc
	IsInstanceReady(context.Context, string, string) wait.ConditionWithContextFunc
}

type btpClient struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &btpClient{
		dynamic: dynamic,
	}
}

func (c *btpClient) GetServiceInstance(ctx context.Context, namespace, name string) (*ServiceInstance, error) {
	obj, err := c.dynamic.Resource(GVRServiceInstance).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	instance := &ServiceInstance{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, instance)

	return instance, err
}

func (c *btpClient) CreateServiceInstance(ctx context.Context, obj *ServiceInstance) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = c.dynamic.Resource(GVRServiceInstance).
		Namespace(obj.GetNamespace()).
		Create(ctx, &unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	return err
}

func (c *btpClient) GetServiceBinding(ctx context.Context, namespace, name string) (*ServiceBinding, error) {
	obj, err := c.dynamic.Resource(GVRServiceBinding).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	binding := &ServiceBinding{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, binding)

	return binding, err
}

func (c *btpClient) CreateServiceBinding(ctx context.Context, obj *ServiceBinding) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = c.dynamic.Resource(GVRServiceBinding).
		Namespace(obj.GetNamespace()).
		Create(ctx, &unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	return err
}

func (c *btpClient) IsBindingReady(ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := c.GetServiceBinding(ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance.Status)
	}
}

func (c *btpClient) IsInstanceReady(ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := c.GetServiceInstance(ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance.Status)
	}
}

func isResourceReady(status CommonStatus) (bool, error) {
	failed := status.IsFailed()
	if failed {
		failedMessage := status.GetConditionMessage("Failed")

		return false, errors.New(failedMessage)
	}

	return status.IsReady(), nil
}
