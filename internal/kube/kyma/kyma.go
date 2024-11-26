package kyma

import (
	"context"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	defaultKymaName      = "default"
	defaultKymaNamespace = "kyma-system"
)

type Interface interface {
	ListModuleReleaseMeta(context.Context) (*ModuleReleaseMetaList, error)
	ListModuleTemplate(context.Context) (*ModuleTemplateList, error)
	GetDefaultKyma(context.Context) (*Kyma, error)
	UpdateDefaultKyma(context.Context, *Kyma) error
	EnableModule(context.Context, string, string) error
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

func (c *client) ListModuleReleaseMeta(ctx context.Context) (*ModuleReleaseMetaList, error) {
	return list[ModuleReleaseMetaList](ctx, c.dynamic, GVRModuleReleaseMeta)
}

func (c *client) ListModuleTemplate(ctx context.Context) (*ModuleTemplateList, error) {
	return list[ModuleTemplateList](ctx, c.dynamic, GVRModuleTemplate)
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

// GetDefaultKyma gets the default Kyma CR from the kyma-system namespace and cast it to the Kyma structure
func (c *client) GetDefaultKyma(ctx context.Context) (*Kyma, error) {
	u, err := c.dynamic.Resource(GVRKyma).
		Namespace(defaultKymaNamespace).
		Get(ctx, defaultKymaName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	kyma := &Kyma{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)

	return kyma, err
}

// UpdateDefaultKyma updates the default Kyma CR from the kyma-system namespace based on the Kyma CR from arguments
func (c *client) UpdateDefaultKyma(ctx context.Context, obj *Kyma) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = c.dynamic.Resource(GVRKyma).
		Namespace(defaultKymaNamespace).
		Update(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})

	return err
}

// EnableModule adds module to the default Kyma CR in the kyma-system namespace
// if moduleChannel is empty it uses default channel in the Kyma CR
func (c *client) EnableModule(ctx context.Context, moduleName, moduleChannel string) error {
	kymaCR, err := c.GetDefaultKyma(ctx)
	if err != nil {
		return err
	}

	kymaCR = enableModule(kymaCR, moduleName, moduleChannel)

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

func enableModule(kymaCR *Kyma, moduleName, moduleChannel string) *Kyma {
	for i, m := range kymaCR.Spec.Modules {
		if m.Name == moduleName {
			// module already exists, update channel
			kymaCR.Spec.Modules[i].Channel = moduleChannel
			return kymaCR
		}
	}

	kymaCR.Spec.Modules = append(kymaCR.Spec.Modules, Module{
		Name:    moduleName,
		Channel: moduleChannel,
	})

	return kymaCR
}

func disableModule(kymaCR *Kyma, moduleName string) *Kyma {
	kymaCR.Spec.Modules = slices.DeleteFunc(kymaCR.Spec.Modules, func(m Module) bool {
		return m.Name == moduleName
	})

	return kymaCR
}
