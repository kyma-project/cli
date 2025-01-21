package kyma

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"k8s.io/utils/ptr"

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
	GetModuleReleaseMetaForModule(context.Context, string) (*ModuleReleaseMeta, error)
	GetModuleTemplateForModule(context.Context, string, string, string) (*ModuleTemplate, error)
	GetDefaultKyma(context.Context) (*Kyma, error)
	UpdateDefaultKyma(context.Context, *Kyma) error
	GetModuleInfo(context.Context, string) (*KymaModuleInfo, error)
	WaitForModuleState(context.Context, string, ...string) error
	EnableModule(context.Context, string, string, string) error
	DisableModule(context.Context, string) error
	ManageModule(context.Context, string, string) error
	UnmanageModule(context.Context, string) error
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

// GetModuleReleaseMetaForModule returns right ModuleReleaseMeta CR corelated with given module name
func (c *client) GetModuleReleaseMetaForModule(ctx context.Context, moduleName string) (*ModuleReleaseMeta, error) {
	list, err := c.ListModuleReleaseMeta(ctx)
	if err != nil {
		return nil, err
	}

	for _, releaseMeta := range list.Items {
		if releaseMeta.Spec.ModuleName == moduleName {
			return &releaseMeta, nil
		}
	}

	return nil, fmt.Errorf("can't find ModuleReleaseMeta CR for module %s", moduleName)
}

// GetModuleTemplateForModule returns ModuleTemplate CR corelated with given module name in right version
func (c *client) GetModuleTemplateForModule(ctx context.Context, moduleName, moduleVersion, moduleChannel string) (*ModuleTemplate, error) {
	moduleTemplates, err := c.ListModuleTemplate(ctx)
	if err != nil {
		return nil, err
	}

	for _, moduleTemplate := range moduleTemplates.Items {
		if moduleTemplate.ObjectMeta.Name == fmt.Sprintf("%s-%s", moduleName, moduleChannel) {
			// old module template detected
			// TODO: tests
			return &moduleTemplate, nil
		}
		// in case this ever stops working we could get moduleReleaseMeta list and parse that
		// https://github.com/kyma-project/cli/issues/2319#issuecomment-2602751723
		if moduleTemplate.Spec.ModuleName == moduleName &&
			moduleTemplate.Spec.Version == moduleVersion {
			return &moduleTemplate, nil
		}
	}

	return nil, fmt.Errorf("can't find ModuleTemplate CR for module %s in version %s", moduleName, moduleVersion)
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

// ManageModule configures given module as managed and updates Kyma CR in the kyma-system namespace with it
func (c *client) ManageModule(ctx context.Context, moduleName, policy string) error {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return err
	}

	kymaCR, err = manageModule(kymaCR, moduleName, policy)
	if err != nil {
		return err
	}

	return c.UpdateDefaultKyma(ctx, kymaCR)
}

// UnmanageModule configures given module as unmanaged and updates Kyma CR in the kyma-system namespace with it
func (c *client) UnmanageModule(ctx context.Context, moduleName string) error {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return err
	}

	kymaCR, err = unmanageModule(kymaCR, moduleName)
	if err != nil {
		return err
	}

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

func manageModule(kymaCR *Kyma, moduleName, policy string) (*Kyma, error) {
	for i, m := range kymaCR.Spec.Modules {
		if m.Name == moduleName {
			// module exists, update managed
			kymaCR.Spec.Modules[i].Managed = ptr.To(true)
			kymaCR.Spec.Modules[i].CustomResourcePolicy = policy

			return kymaCR, nil
		}
	}

	return kymaCR, errors.New("module not found")
}

func unmanageModule(kymaCR *Kyma, moduleName string) (*Kyma, error) {
	for i, m := range kymaCR.Spec.Modules {
		if m.Name == moduleName {
			// module exists, update managed
			kymaCR.Spec.Modules[i].Managed = ptr.To(false)
			kymaCR.Spec.Modules[i].CustomResourcePolicy = "Ignore"

			return kymaCR, nil
		}
	}

	return kymaCR, errors.New("module not found")
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
