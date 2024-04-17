package kube

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type somethingWithStatus struct {
	Status Status
}

type Status struct {
	Conditions []metav1.Condition
	Ready      string
	InstanceID string
}

func GetServiceStatus(u *unstructured.Unstructured) (Status, error) {
	instance := somethingWithStatus{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &instance); err != nil {
		return Status{}, &clierror.Error{
			Message: "failed to read resource data",
			Details: err.Error(),
		}
	}

	return instance.Status, nil
}

// IsReady returns readiness status
func IsReady(u *unstructured.Unstructured) (bool, error) {
	status, err := GetServiceStatus(u)
	if err != nil {
		return false, err
	}

	ready := (status.Ready == "True") &&
		isConditionTrue(status.Conditions, "Succeeded") &&
		isConditionTrue(status.Conditions, "Ready")
	return ready, nil
}

// IsFailed returns if at least one condition has Failed status
func IsFailed(u *unstructured.Unstructured) (bool, error) {
	status, err := GetServiceStatus(u)
	if err != nil {
		return true, err
	}

	failed := (status.Ready == "False") &&
		isConditionTrue(status.Conditions, "Failed")
	return failed, nil
}

func isConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	condition := meta.FindStatusCondition(conditions, conditionType)
	return condition != nil && condition.Status == metav1.ConditionTrue
}

func GetServiceInstance(kubeClient Client, ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return kubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func GetServiceBinding(kubeClient Client, ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func GetConditionMessage(conditions []metav1.Condition, conditionType string) string {
	condition := meta.FindStatusCondition(conditions, conditionType)
	if condition == nil {
		return ""
	}
	return condition.Message
}

func isResourceReady(instance *unstructured.Unstructured) (bool, error) {
	failed, err := IsFailed(instance)
	if err != nil {
		return false, err
	}
	if failed {
		status, err := GetServiceStatus(instance)
		if err != nil {
			return false, err
		}
		failedMessage := GetConditionMessage(status.Conditions, "Failed")

		return false, &clierror.Error{
			Message: failedMessage,
		}

	}

	ready, err := IsReady(instance)
	if err != nil {
		return ready, err
	}
	return ready, nil
}

func IsBindingReady(kubeClient Client, ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := GetServiceBinding(kubeClient, ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance)
	}
}

func IsInstanceReady(kubeClient Client, ctx context.Context, namespace, name string) wait.ConditionWithContextFunc {
	return func(_ context.Context) (bool, error) {
		instance, err := GetServiceInstance(kubeClient, ctx, namespace, name)
		if err != nil {
			return false, err
		}
		return isResourceReady(instance)
	}
}
