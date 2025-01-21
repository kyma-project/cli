package kyma

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

func TestGetDefaultKyma(t *testing.T) {
	t.Run("get Kyma from the cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme, fixDefaultKyma()))

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name: "test-module",
					},
				},
			},
			Status: KymaStatus{
				Modules: []ModuleStatus{
					{
						Name:    "test-module",
						Version: "1.2.3",
						State:   "Ready",
					},
				},
			},
		}

		kyma, err := client.GetDefaultKyma(context.Background())
		require.NoError(t, err)
		require.Equal(t, expectedKyma, kyma)
	})

	t.Run("kyma not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme))

		kyma, err := client.GetDefaultKyma(context.Background())
		require.ErrorContains(t, err, "not found")
		require.Nil(t, kyma)
	})
}

func TestUpdateDefaultKyma(t *testing.T) {
	t.Run("update kyma", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, fixDefaultKyma())
		client := NewClient(dynamic)

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name:    "test-module",
						Channel: "regular",
					},
					{
						Name: "another-test-module",
					},
				},
			},
		}

		err := client.UpdateDefaultKyma(context.Background(), expectedKyma)
		require.NoError(t, err)

		u, err := dynamic.Resource(GVRKyma).Namespace("kyma-system").Get(context.Background(), "default", v1.GetOptions{})
		require.NoError(t, err)

		kyma := &Kyma{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)
		require.NoError(t, err)
		require.Equal(t, expectedKyma, kyma)
	})

	t.Run("kyma not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme))

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name:    "test-module",
						Channel: "regular",
					},
					{
						Name: "another-test-module",
					},
				},
			},
		}

		err := client.UpdateDefaultKyma(context.Background(), expectedKyma)
		require.ErrorContains(t, err, "not found")
	})
}

func Test_disableModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kymaCR     *Kyma
		moduleName string
		want       *Kyma
	}{
		{
			name: "unchanged modules list",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
					},
				},
			},
			moduleName: "module",
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
					},
				},
			},
		},
		{
			name: "changed modules list",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
						{
							Name: "module",
						},
					},
				},
			},
			moduleName: "module",
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := disableModule(kymaCR, moduleName)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			var gotInterface map[string]interface{}
			var wantInterface map[string]interface{}
			err = json.Unmarshal(gotBytes, &gotInterface)
			require.NoError(t, err)
			err = json.Unmarshal(wantBytes, &wantInterface)

			require.NoError(t, err)
			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("updateCR() = %v, want %v", gotInterface, wantInterface)
			}
		})
	}
}

func Test_enableModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		kymaCR               *Kyma
		moduleName           string
		customResourcePolicy string
		channel              string
		want                 *Kyma
	}{
		{
			name:                 "unchanged modules list",
			moduleName:           "module",
			channel:              "",
			customResourcePolicy: "",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "module",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "module",
						},
					},
				},
			},
		},
		{
			name:       "added module",
			moduleName: "module",
			channel:    "",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
						{
							Name: "module",
						},
					},
				},
			},
		},
		{
			name:                 "added module with channel and customResourcePolicy",
			moduleName:           "module",
			channel:              "channel",
			customResourcePolicy: "customResourcePolicy",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "istio",
						},
						{
							Name:                 "module",
							Channel:              "channel",
							CustomResourcePolicy: "customResourcePolicy",
						},
					},
				},
			},
		},
		{
			name:                 "added channel and customResourcePolicy to existing module",
			moduleName:           "module",
			channel:              "channel",
			customResourcePolicy: "customResourcePolicy",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "module",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Channel:              "channel",
							CustomResourcePolicy: "customResourcePolicy",
						},
					},
				},
			},
		},
		{
			name:       "removed channel from existing module",
			moduleName: "module",
			channel:    "",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:    "module",
							Channel: "channel",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name: "module",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		moduleChannel := tt.channel
		customResourcePolicy := tt.customResourcePolicy
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := enableModule(kymaCR, moduleName, moduleChannel, customResourcePolicy)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			var gotInterface map[string]interface{}
			var wantInterface map[string]interface{}
			err = json.Unmarshal(gotBytes, &gotInterface)
			require.NoError(t, err)
			err = json.Unmarshal(wantBytes, &wantInterface)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("updateCR() = %v, want %v", gotInterface, wantInterface)
			}
		})
	}
}

func fixDefaultKyma() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]interface{}{
				"name":      "default",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"channel": "fast",
				"modules": []interface{}{
					map[string]interface{}{
						"name": "test-module",
					},
				},
			},
			"status": map[string]interface{}{
				"modules": []interface{}{
					map[string]interface{}{
						"name":    "test-module",
						"version": "1.2.3",
						"state":   "Ready",
					},
				},
			},
		},
	}
}

func Test_client_ListModuleReleaseMeta(t *testing.T) {
	t.Run("list ModuleReleaseMeta", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleReleaseMeta.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleReleaseMeta("test-1"),
			fixModuleReleaseMeta("test-2"),
		))

		list, err := client.ListModuleReleaseMeta(context.Background())

		require.NoError(t, err)
		require.Len(t, list.Items, 2)
		require.Contains(t, list.Items, fixModuleReleaseMetaStruct("test-1"))
		require.Contains(t, list.Items, fixModuleReleaseMetaStruct("test-2"))

	})
}

func Test_client_ListModuleTemplate(t *testing.T) {
	t.Run("list ModuleTemplate", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleTemplate.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleTemplate("test-1", "0.1", ""),
			fixModuleTemplate("test-2", "0.1", ""),
		))

		list, err := client.ListModuleTemplate(context.Background())

		require.NoError(t, err)
		require.Len(t, list.Items, 2)
		require.Contains(t, list.Items, fixModuleTemplateStruct("test-1", "0.1", ""))
		require.Contains(t, list.Items, fixModuleTemplateStruct("test-2", "0.1", ""))

	})
}

func Test_client_GetModuleReleaseMetaForModule(t *testing.T) {
	t.Run("get module ModuleReleaseMeta", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleReleaseMeta.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleReleaseMeta("test-1"),
			fixModuleReleaseMeta("test-2"),
		))

		got, err := client.GetModuleReleaseMetaForModule(context.Background(), "test-2")

		require.NoError(t, err)
		require.Equal(t, fixModuleReleaseMetaStruct("test-2"), *got)
	})

	t.Run("no ModuleReleaseMeta for module", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleReleaseMeta.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleReleaseMeta("test-1"),
			fixModuleReleaseMeta("test-2"),
		))

		got, err := client.GetModuleReleaseMetaForModule(context.Background(), "test-123")

		require.ErrorContains(t, err, "can't find ModuleReleaseMeta CR for module test-123")
		require.Nil(t, got)
	})
}

func Test_client_GetModuleTemplateForModule(t *testing.T) {
	t.Run("get module ModuleTemplate", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleTemplate.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleTemplate("test-1", "0.1", ""),
			fixModuleTemplate("test-2", "0.1", ""),
		))

		got, err := client.GetModuleTemplateForModule(context.Background(), "test-2", "0.1", "")

		require.NoError(t, err)
		require.Equal(t, fixModuleTemplateStruct("test-2", "0.1", ""), *got)
	})

	t.Run("get module ModuleTemplate for old module", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleTemplate.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleTemplate("test-1", "0.1", ""),
			fixModuleTemplate("test-2", "0.1", ""),
			fixModuleTemplate("test", "", "fast"),
		))

		got, err := client.GetModuleTemplateForModule(context.Background(), "test", "0.1", "fast")

		require.NoError(t, err)
		require.Equal(t, fixModuleTemplateStruct("test", "", "fast"), *got)
	})

	t.Run("no ModuleReleaseMeta for module", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRModuleTemplate.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme,
			fixModuleTemplate("test-1", "0.1", ""),
			fixModuleTemplate("test-2", "0.1", ""),
		))

		got, err := client.GetModuleTemplateForModule(context.Background(), "test-2", "0.2", "")

		require.ErrorContains(t, err, "can't find ModuleTemplate CR for module test-2 in version 0.2")
		require.Nil(t, got)
	})
}

func Test_client_GetModuleInfo(t *testing.T) {
	t.Run("get ModuleInfo", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme, fixDefaultKyma()))

		got, err := client.GetModuleInfo(context.Background(), "test-module")
		require.NoError(t, err)
		require.Equal(t, KymaModuleInfo{
			Spec: Module{
				Name: "test-module",
			},
			Status: ModuleStatus{
				Name:    "test-module",
				Version: "1.2.3",
				State:   "Ready",
			},
		}, *got)
	})

	t.Run("get error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		client := NewClient(dynamic_fake.NewSimpleDynamicClient(scheme))

		got, err := client.GetModuleInfo(context.Background(), "test-module")
		require.ErrorContains(t, err, "not found")
		require.Nil(t, got)
	})
}

func fixModuleReleaseMetaStruct(moduleName string) ModuleReleaseMeta {
	return ModuleReleaseMeta{
		TypeMeta: v1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleReleaseMeta",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      moduleName,
			Namespace: "kyma-system",
		},
		Spec: ModuleReleaseMetaSpec{
			ModuleName: moduleName,
			Channels: []ChannelVersionAssignment{
				{
					Version: "0.1",
					Channel: "regular",
				},
				{
					Version: "0.2",
					Channel: "fast",
				},
			},
		},
	}
}

func fixModuleTemplateStruct(moduleName, moduleVersion, moduleChannel string) ModuleTemplate {
	mt := ModuleTemplate{
		TypeMeta: v1.TypeMeta{
			APIVersion: "operator.kyma-project.io/v1beta2",
			Kind:       "ModuleTemplate",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      moduleName,
			Namespace: "kyma-system",
		},
		Spec: ModuleTemplateSpec{
			ModuleName: moduleName,
		},
	}
	if moduleVersion != "" {
		mt.Spec.Version = moduleVersion
	}
	if moduleChannel != "" {
		mt.ObjectMeta.Name = fmt.Sprintf("%s-%s", moduleName, moduleChannel)
		mt.Spec.Channel = moduleChannel
	}

	return mt
}

func fixModuleReleaseMeta(moduleName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]interface{}{
				"name":      moduleName,
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": moduleName,
				"channels": []interface{}{
					map[string]interface{}{
						"version": "0.1",
						"channel": "regular",
					},
					map[string]interface{}{
						"version": "0.2",
						"channel": "fast",
					},
				},
			},
		},
	}
}

func fixModuleTemplate(moduleName, moduleVersion, moduleChannel string) *unstructured.Unstructured {
	mt := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      moduleName,
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": moduleName,
			},
		},
	}

	if moduleVersion != "" {
		_ = unstructured.SetNestedField(mt.Object, moduleVersion, "spec", "version")
	}
	if moduleChannel != "" {
		_ = unstructured.SetNestedField(mt.Object, fmt.Sprintf("%s-%s", moduleName, moduleChannel), "metadata", "name")
		_ = unstructured.SetNestedField(mt.Object, moduleChannel, "spec", "channel")
	}

	return mt
}

func Test_manageModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kymaCR     *Kyma
		moduleName string
		policy     string
		want       *Kyma
	}{
		{
			name:       "unchanged module",
			moduleName: "module",
			policy:     "CreateAndDelete",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "CreateAndDelete",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "CreateAndDelete",
						},
					},
				},
			},
		},
		{
			name:       "already managed, configuration changed",
			moduleName: "module",
			policy:     "Ignore",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "CreateAndDelete",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "Ignore",
						},
					},
				},
			},
		},
		{
			name:       "module updated",
			moduleName: "module",
			policy:     "CreateAndDelete",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(false),
							CustomResourcePolicy: "Ignore",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "CreateAndDelete",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		want := tt.want
		policy := tt.policy
		t.Run(tt.name, func(t *testing.T) {
			got, err := manageModule(kymaCR, moduleName, policy)
			require.NoError(t, err)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			var gotInterface map[string]interface{}
			var wantInterface map[string]interface{}
			err = json.Unmarshal(gotBytes, &gotInterface)
			require.NoError(t, err)
			err = json.Unmarshal(wantBytes, &wantInterface)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("updateCR() = %v, want %v", gotInterface, wantInterface)
			}
		})
	}
}

func Test_unmanageModule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kymaCR     *Kyma
		moduleName string
		want       *Kyma
	}{
		{
			name:       "unchanged module",
			moduleName: "module",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(false),
							CustomResourcePolicy: "Ignore",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(false),
							CustomResourcePolicy: "Ignore",
						},
					},
				},
			},
		},
		{
			name:       "module updated",
			moduleName: "module",
			kymaCR: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(true),
							CustomResourcePolicy: "CreateAndDelete",
						},
					},
				},
			},
			want: &Kyma{
				Spec: KymaSpec{
					Modules: []Module{
						{
							Name:                 "module",
							Managed:              ptr.To(false),
							CustomResourcePolicy: "Ignore",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kymaCR := tt.kymaCR
		moduleName := tt.moduleName
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmanageModule(kymaCR, moduleName)
			require.NoError(t, err)
			gotBytes, err := json.Marshal(got)
			require.NoError(t, err)
			wantBytes, err := json.Marshal(want)
			require.NoError(t, err)
			var gotInterface map[string]interface{}
			var wantInterface map[string]interface{}
			err = json.Unmarshal(gotBytes, &gotInterface)
			require.NoError(t, err)
			err = json.Unmarshal(wantBytes, &wantInterface)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("updateCR() = %v, want %v", gotInterface, wantInterface)
			}
		})
	}
}
