package kyma

import (
	"context"
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	DefaultKymaName                     = "default"
	DefaultKymaNamespace                = "kyma-system"
	CustomResourcePolicyIgnore          = "Ignore"
	CustomResourcePolicyCreateAndDelete = "CreateAndDelete"
)

type Interface interface {
	ListModuleReleaseMeta(context.Context) (*ModuleReleaseMetaList, error)
	ListModuleTemplate(context.Context) (*ModuleTemplateList, error)
	GetDefaultKyma(context.Context) (*Kyma, error)
	UpdateDefaultKyma(context.Context, *Kyma) error
	WaitForModuleState(context.Context, string, ...string) error
	GetModuleInfo(context.Context, string) (*KymaModuleInfo, error)
	EnableModule(context.Context, string, string, string) error
	DisableModule(context.Context, string) error
}

type client struct {
	dynamic dynamic.Interface
}

func NewClient(dynamic dynamic.Interface) Interface {
	return &client{
		dynamic: dynamic,
	}
}

// ListModuleReleaseMeta lists ModuleReleaseMeta resources from across the whole cluster
func (c *client) ListModuleReleaseMeta(ctx context.Context) (*ModuleReleaseMetaList, error) {
	return list[ModuleReleaseMetaList](ctx, c.dynamic, GVRModuleReleaseMeta)
}

// ListModuleTemplate lists ModuleTemplate resources from across the whole cluster
func (c *client) ListModuleTemplate(ctx context.Context) (*ModuleTemplateList, error) {
	return list[ModuleTemplateList](ctx, c.dynamic, GVRModuleTemplate)
}

// GetDefaultKyma gets the default Kyma CR from the kyma-system namespace and cast it to the Kyma structure
func (c *client) GetDefaultKyma(ctx context.Context) (*Kyma, error) {
	u, err := c.dynamic.Resource(GVRKyma).
		Namespace(DefaultKymaNamespace).
		Get(ctx, DefaultKymaName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	kyma := &Kyma{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)

	return kyma, err
}

// GetModuleState returns state of the specific module based on the default kyma on the cluster
func (c *client) GetModuleInfo(ctx context.Context, moduleName string) (*KymaModuleInfo, error) {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return nil, err
	}

	return getmoduleInfo(kymaCR, moduleName), nil
}

// UpdateDefaultKyma updates the default Kyma CR from the kyma-system namespace based on the Kyma CR from arguments
func (c *client) UpdateDefaultKyma(ctx context.Context, obj *Kyma) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = c.dynamic.Resource(GVRKyma).
		Namespace(DefaultKymaNamespace).
		Update(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})

	return err
}

// WaitForModuleState waits until module is not in on of given expected states
func (c *client) WaitForModuleState(ctx context.Context, moduleName string, expectedStates ...string) error {
	watcher, err := c.dynamic.Resource(GVRKyma).
		Namespace(DefaultKymaNamespace).
		Watch(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", DefaultKymaName),
		})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s with last error: %s", ctx.Err(), lastErr)
		case event := <-watcher.ResultChan():
			err = checkModuleState(event.Object, moduleName, expectedStates...)
			if err != nil {
				// set last error and try one more time
				lastErr = err
				continue
			}

			return nil
		}
	}
}

// EnableModule adds module to the default Kyma CR in the kyma-system namespace
// if moduleChannel is empty it uses default channel in the Kyma CR
func (c *client) EnableModule(ctx context.Context, moduleName, moduleChannel, customResourcePolicy string) error {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return err
	}

	kymaCR = enableModule(kymaCR, moduleName, moduleChannel, customResourcePolicy)

	return c.UpdateDefaultKyma(ctx, kymaCR)
}

// DisableModule removes module from the default Kyma CR in the kyma-system namespace
func (c *client) DisableModule(ctx context.Context, moduleName string) error {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return err
	}

	kymaCR = disableModule(kymaCR, moduleName)

	return c.UpdateDefaultKyma(ctx, kymaCR)
}

func checkModuleState(kymaObj runtime.Object, moduleName string, expectedStates ...string) error {
	kyma := &Kyma{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(kymaObj.(*unstructured.Unstructured).Object, kyma)
	if err != nil {
		return err
	}

	moduleInfo := getmoduleInfo(kyma, moduleName)

	for _, expectedState := range expectedStates {
		if moduleInfo.Status.State == expectedState {
			return nil
		}
	}

	return fmt.Errorf("module %s is in the %s state", moduleName, moduleInfo.Status.State)
}

func getmoduleInfo(kymaCR *Kyma, moduleName string) *KymaModuleInfo {
	info := KymaModuleInfo{}
	for _, module := range kymaCR.Spec.Modules {
		if module.Name == moduleName {
			info.Spec = module
		}
	}

	for _, module := range kymaCR.Status.Modules {
		if module.Name == moduleName {
			info.Status = module
		}
	}

	return &info
}

func enableModule(kymaCR *Kyma, moduleName, moduleChannel, customResourcePolicy string) *Kyma {
	for i, m := range kymaCR.Spec.Modules {
		if m.Name == moduleName {
			// module already exists, update channel
			kymaCR.Spec.Modules[i].Channel = moduleChannel
			kymaCR.Spec.Modules[i].CustomResourcePolicy = customResourcePolicy
			return kymaCR
		}
	}

	kymaCR.Spec.Modules = append(kymaCR.Spec.Modules, Module{
		Name:                 moduleName,
		Channel:              moduleChannel,
		CustomResourcePolicy: customResourcePolicy,
	})

	return kymaCR
}

func disableModule(kymaCR *Kyma, moduleName string) *Kyma {
	kymaCR.Spec.Modules = slices.DeleteFunc(kymaCR.Spec.Modules, func(m Module) bool {
		return m.Name == moduleName
	})

	return kymaCR
}

func list[T any](ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource) (*T, error) {
	list, err := client.Resource(gvr).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	structuredList := new(T)
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(list.UnstructuredContent(), structuredList)
	return structuredList, err
}
