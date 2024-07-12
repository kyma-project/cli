package btp

import (
	"context"
	"errors"

	"github.com/kyma-project/cli.v3/internal/kube"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

// IsReady returns readiness status
func IsReady(status CommonStatus) bool {
	return (status.Ready == "True") &&
		isConditionTrue(status.Conditions, "Succeeded") &&
		isConditionTrue(status.Conditions, "Ready")
}

// IsFailed returns if at least one condition has Failed status
func IsFailed(status CommonStatus) bool {
	return (status.Ready == "False") &&
		isConditionTrue(status.Conditions, "Failed")
}

func isConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	condition := meta.FindStatusCondition(conditions, conditionType)
	return condition != nil && condition.Status == metav1.ConditionTrue
}

func GetServiceInstance(kubeClient kube.Client, ctx context.Context, namespace, name string) (*ServiceInstance, error) {
	obj, err := kubeClient.Dynamic().Resource(GVRServiceInstance).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	instance := &ServiceInstance{}
	err = kube.FromUnstructured(obj, instance)

	return instance, err
}

func CreateServiceInstance(kubeClient kube.Client, ctx context.Context, obj *ServiceInstance) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = kubeClient.Dynamic().Resource(GVRServiceInstance).
		Namespace(obj.GetNamespace()).
		Create(ctx, &unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	return err
}

func GetServiceBinding(kubeClient kube.Client, ctx context.Context, namespace, name string) (*ServiceBinding, error) {
	obj, err := kubeClient.Dynamic().Resource(GVRServiceBinding).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	binding := &ServiceBinding{}
	err = kube.FromUnstructured(obj, binding)

	return binding, err
}

func CreateServiceBinding(kubeClient kube.Client, ctx context.Context, obj *ServiceBinding) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = kubeClient.Dynamic().Resource(GVRServiceBinding).
		Namespace(obj.GetNamespace()).
		Create(ctx, &unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	return err
}

func GetConditionMessage(conditions []metav1.Condition, conditionType string) string {
	condition := meta.FindStatusCondition(conditions, conditionType)
	if condition == nil {
		return ""
	}
	return condition.Message
}

func isResourceReady(status CommonStatus) (bool, error) {
	failed := IsFailed(status)
	if failed {
		failedMessage := GetConditionMessage(status.Conditions, "Failed")

		return false, errors.New(failedMessage)
	}

	return IsReady(status), nil
}

func IsBindingReady(kubeClient kube.Client, ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := GetServiceBinding(kubeClient, ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance.Status)
	}
}

func IsInstanceReady(kubeClient kube.Client, ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := GetServiceInstance(kubeClient, ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance.Status)
	}
}
