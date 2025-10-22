package diagnostics_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
)

func TestNewModuleLogsCollector(t *testing.T) {
	// Given
	kubeClient := &kube_fake.KubeClient{
		TestKubernetesInterface: fake.NewSimpleClientset(),
	}
	var writer bytes.Buffer
	verbose := true

	// When
	collector := diagnostics.NewModuleLogsCollector(kubeClient, &writer, verbose)

	// Then
	assert.NotNil(t, collector)
}

func TestRunSince_EmptyModules(t *testing.T) {
	// Given
	kubeClient := &kube_fake.KubeClient{
		TestKubernetesInterface: fake.NewSimpleClientset(),
	}
	var writer bytes.Buffer
	collector := diagnostics.NewModuleLogsCollector(kubeClient, &writer, false)

	ctx := context.Background()
	modules := []string{}
	since := 10 * time.Minute

	// When
	result := collector.RunSince(ctx, modules, since)

	// Then
	assert.NotNil(t, result)
	assert.Empty(t, result.Logs)
}

func TestRunLast_EmptyModules(t *testing.T) {
	// Given
	kubeClient := &kube_fake.KubeClient{
		TestKubernetesInterface: fake.NewSimpleClientset(),
	}
	var writer bytes.Buffer
	collector := diagnostics.NewModuleLogsCollector(kubeClient, &writer, false)

	ctx := context.Background()
	modules := []string{}
	lines := int64(100)

	// When
	result := collector.RunLast(ctx, modules, lines)

	// Then
	assert.NotNil(t, result)
	assert.Empty(t, result.Logs)
}

func TestRunSince_MultipleModules(t *testing.T) {
	// Given
	pod1 := createTestPod("test-pod-1", "kyma-system", "module-1")
	pod2 := createTestPod("test-pod-2", "kyma-system", "module-2")
	kubeClient := createMockKubeClient([]runtime.Object{&pod1, &pod2})
	var writer bytes.Buffer
	collector := diagnostics.NewModuleLogsCollector(kubeClient, &writer, false)

	ctx := context.Background()
	modules := []string{"module-1", "module-2"}
	since := 10 * time.Minute

	// When
	result := collector.RunSince(ctx, modules, since)

	// Then
	assert.NotNil(t, result)
}

// Helper functions

func createTestPod(name, namespace, module string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"kyma-project.io/module": module,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}

func createMockKubeClient(objects []runtime.Object) *kube_fake.KubeClient {
	return &kube_fake.KubeClient{
		TestKubernetesInterface: fake.NewSimpleClientset(objects...),
	}
}
