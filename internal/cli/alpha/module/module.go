package module

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var kymaResource = schema.GroupVersionResource{
	Group:    "operator.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "kymas",
}

type ModulesList []interface{}

type Interactor struct {
	Client       kube.KymaKube
	ResourceName types.NamespacedName
	ForceUpdate  bool
}

func NewInteractor(client kube.KymaKube, name types.NamespacedName, forceUpdate bool) Interactor {
	return Interactor{
		Client:       client,
		ResourceName: name,
		ForceUpdate:  forceUpdate,
	}
}

func (i *Interactor) Get(ctx context.Context) (ModulesList, error) {
	namespace := i.ResourceName.Namespace
	name := i.ResourceName.Name
	kyma, err := i.Client.Dynamic().Resource(kymaResource).Namespace(namespace).Get(
		ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get Kyma with name %s and namespace %s: %w", name, namespace, err)
	}

	modules, _, err := unstructured.NestedSlice(kyma.UnstructuredContent(), "spec", "modules")
	if err != nil {
		return nil, fmt.Errorf("could not parse modules spec: %w", err)
	}
	return modules, nil
}

func (i *Interactor) Update(ctx context.Context, modules ModulesList) error {
	namespace := i.ResourceName.Namespace
	name := i.ResourceName.Name
	kyma, err := i.Client.Dynamic().Resource(kymaResource).Namespace(namespace).Get(
		ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get Kyma %s/%s: %w", namespace, name, err)
	}

	err = unstructured.SetNestedSlice(kyma.Object, modules, "spec", "modules")
	if err != nil {
		return fmt.Errorf("failed to set modules list in Kyma spec: %w", err)
	}
	kyma.SetManagedFields(nil)
	data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, kyma)
	if err != nil {
		return fmt.Errorf("failed to marshal kyma object: %w", err)
	}
	patchOpts := metav1.PatchOptions{FieldManager: "kyma"}
	patchOpts.Force = pointer.Bool(i.ForceUpdate)
	_, err = i.Client.Dynamic().Resource(kymaResource).Namespace(namespace).Patch(
		ctx, kyma.GetName(), types.ApplyPatchType, data, patchOpts)
	if err != nil {
		return fmt.Errorf("failed to update Kyma %s in %s: %w", name, namespace, err)
	}
	return nil
}

func (i *Interactor) WaitForKymaReadiness() error {
	namespace := i.ResourceName.Namespace
	name := i.ResourceName.Name
	time.Sleep(2 * time.Second)
	checkFn := func(u *unstructured.Unstructured) (bool, error) {
		status, _, err := unstructured.NestedString(u.Object, "status", "state")
		if err != nil {
			return false, errors.Wrap(err, "could not parse status state")
		}

		switch status {
		case "Ready":
			fmt.Println("Modules successfully changed. Kyma in state 'Ready'")
			return true, nil
		case "Error":
			operation, exists, err := unstructured.NestedString(u.Object, "status", "lastOperation", "operation")
			if err != nil {
				return false, fmt.Errorf("could not parse status lastOperation: %w", err)
			}
			if exists {
				return false, errors.Errorf("updating Kyma resulted in error state with last operation: %s", operation)
			}
			return false, errors.New("Kyma in state error")
		default:
			return false, nil
		}
	}
	err := i.Client.WatchResource(kymaResource, name, namespace, checkFn)
	if err != nil {
		return err
	}
	return nil
}
