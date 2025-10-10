package diagnostics_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func mockMetricsHandler(nodeNames []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, nodeName := range nodeNames {
			if r.URL.Path == "/apis/metrics.k8s.io/v1beta1/nodes/"+nodeName {
				metrics := diagnostics.NodeMetrics{
					Usage: struct {
						CPU    string `json:"cpu"`
						Memory string `json:"memory"`
					}{
						CPU:    "213000000n", // 213 millicores
						Memory: "1073741824", // 1GiB in bytes
					},
				}

				w.Header().Set("Content-Type", "application/json")
				jsonEncoder := json.NewEncoder(w)
				_ = jsonEncoder.Encode(metrics)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})
}

func createMockKubeClientWithRESTClient(nodeNames []string) *fake.KubeClient {
	fakeKubeClient := kubefake.NewSimpleClientset()

	server := httptest.NewServer(mockMetricsHandler(nodeNames))

	restConfig := &rest.Config{
		Host:    server.URL,
		APIPath: "/",
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &schema.GroupVersion{Group: "", Version: "v1"},
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
	}

	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		fmt.Printf("Error creating REST client: %v\n", err)
		return nil
	}

	return &fake.KubeClient{
		TestKubernetesInterface: fakeKubeClient,
		TestRestClient:          restClient,
		TestRestConfig:          restConfig,
	}
}

func setupMockClient(nodes []corev1.Node) *fake.KubeClient {
	var mockClient *fake.KubeClient
	if len(nodes) > 0 {
		var nodeNames []string
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		mockClient = createMockKubeClientWithRESTClient(nodeNames)
	} else {
		fakeKubeClient := kubefake.NewSimpleClientset()
		mockClient = &fake.KubeClient{
			TestKubernetesInterface: fakeKubeClient,
		}
	}

	for _, node := range nodes {
		_, err := mockClient.TestKubernetesInterface.CoreV1().Nodes().Create(context.TODO(), &node, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
	}

	return mockClient
}

func TestNewNodeResourceInfoCollector(t *testing.T) {
	// Given
	fakeKubeClient := kubefake.NewSimpleClientset()
	fakeClient := &fake.KubeClient{
		TestKubernetesInterface: fakeKubeClient,
	}
	var writer bytes.Buffer
	verbose := true

	// When
	collector := diagnostics.NewNodeResourceInfoCollector(fakeClient, &writer, verbose)

	// Then
	assert.NotNil(t, collector)
}

func assertNodeBasicInfo(t *testing.T, result diagnostics.NodeResourceInfo, expectedNode corev1.Node) {
	// Check MachineInfo
	assert.Equal(t, expectedNode.Name, result.MachineInfo.Name)
	assert.Equal(t, expectedNode.Status.NodeInfo.Architecture, result.MachineInfo.Architecture)
	assert.Equal(t, expectedNode.Status.NodeInfo.KernelVersion, result.MachineInfo.KernelVersion)
	assert.Equal(t, expectedNode.Status.NodeInfo.OSImage, result.MachineInfo.OSImage)
	assert.Equal(t, expectedNode.Status.NodeInfo.ContainerRuntimeVersion, result.MachineInfo.ContainerRuntime)
	assert.Equal(t, expectedNode.Status.NodeInfo.KubeletVersion, result.MachineInfo.KubeletVersion)
	assert.Equal(t, expectedNode.Status.NodeInfo.OperatingSystem, result.MachineInfo.OperatingSystem)

	// Check capacity resources
	assert.Equal(t, expectedNode.Status.Capacity.Cpu().String(), result.Capacity.CPU)
	assert.Equal(t, expectedNode.Status.Capacity.Memory().String(), result.Capacity.Memory)
	assert.Equal(t, expectedNode.Status.Capacity.StorageEphemeral().String(), result.Capacity.EphemeralStorage)
	assert.Equal(t, expectedNode.Status.Capacity.Pods().String(), result.Capacity.Pods)
}

func TestRunWithNoNodes(t *testing.T) {
	// Given
	var writer bytes.Buffer
	nodes := []corev1.Node{}
	mockClient := setupMockClient(nodes)
	collector := diagnostics.NewNodeResourceInfoCollector(mockClient, &writer, false)

	// When
	results := collector.Run(context.TODO())

	// Then
	assert.Len(t, results, 0)
	assert.Empty(t, writer.String())
}

func TestRunWithMultipleNodesWithoutPods(t *testing.T) {
	// Given
	var writer bytes.Buffer
	nodes := []corev1.Node{
		createTestNode("node1", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0"),
		createTestNode("node2", "arm64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.1"),
		createTestNode("node3", "amd64", "5.8.0-generic", "Ubuntu 22.04.1 LTS", "containerd://1.6.2", "v1.26.2"),
	}
	mockClient := setupMockClient(nodes)
	collector := diagnostics.NewNodeResourceInfoCollector(mockClient, &writer, false)

	// When
	results := collector.Run(context.TODO())

	// Then
	assert.Len(t, results, 3)

	for i, result := range results {
		expectedNode := nodes[i]
		assertNodeBasicInfo(t, result, expectedNode)

		// All nodes should have no pods
		assert.Equal(t, "0", result.Usage.CPURequests)
		assert.Equal(t, "0", result.Usage.MemoryRequests)
		assert.Equal(t, 0, result.Usage.PodCount)
		assert.Equal(t, "0.0%", result.Usage.CPULimitPercent)
		assert.Equal(t, "0.0%", result.Usage.MemoryLimitPercent)

		// Metrics from mock REST client
		assert.Equal(t, "213.0m", result.Usage.CPUUsage)
		assert.Equal(t, "1024.0Mi", result.Usage.MemoryUsage)
		assert.Equal(t, "5.5%", result.Usage.CPUUsagePercent)
		assert.Equal(t, "14.3%", result.Usage.MemoryUsagePercent)

		// Should have topology labels
		assert.Equal(t, "region", result.Topology.Region)
		assert.Equal(t, "zone", result.Topology.Zone)
	}

	assert.Empty(t, writer.String())
}

func TestRunWithRunningPods(t *testing.T) {
	// Given
	var writer bytes.Buffer
	node := createTestNode("node-with-pods", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0")
	nodes := []corev1.Node{node}
	mockClient := setupMockClient(nodes)

	// Create test pod with resources
	testPod := createTestPodWithResources("test-pod", node.Name)
	_, err := mockClient.TestKubernetesInterface.CoreV1().Pods("default").Create(context.TODO(), &testPod, metav1.CreateOptions{})
	assert.NoError(t, err)

	collector := diagnostics.NewNodeResourceInfoCollector(mockClient, &writer, false)

	// When
	results := collector.Run(context.TODO())

	// Then
	assert.Len(t, results, 1)
	result := results[0]

	assertNodeBasicInfo(t, result, node)

	// Resource totals
	assert.Equal(t, "100m", result.Usage.CPURequests)
	assert.Equal(t, "128Mi", result.Usage.MemoryRequests)
	assert.Equal(t, "200m", result.Usage.CPULimits)
	assert.Equal(t, "256Mi", result.Usage.MemoryLimits)

	// Available resources (allocatable - requests)
	assert.Equal(t, "3800m", result.Usage.CPUAvailable) // 3900m - 100m
	assert.Equal(t, "109", result.Usage.PodsAvailable)  // 110 - 1
	assert.Equal(t, node.Status.Allocatable.StorageEphemeral().String(), result.Usage.EphemeralStorageAvailable)

	// Percentage calculations
	assert.Equal(t, "2.6%", result.Usage.CPURequestedPercent) // 100m / 3900m * 100

	// Pod count
	assert.Equal(t, 1, result.Usage.PodCount)

	// Limit percentages should be calculated
	assert.Equal(t, "5.1%", result.Usage.CPULimitPercent)    // 200m / 3900m * 100
	assert.Equal(t, "3.6%", result.Usage.MemoryLimitPercent) // Should be calculated based on limits

	// Metrics from mock REST client
	assert.Equal(t, "213.0m", result.Usage.CPUUsage)
	assert.Equal(t, "1024.0Mi", result.Usage.MemoryUsage)
	assert.Equal(t, "5.5%", result.Usage.CPUUsagePercent)
	assert.Equal(t, "14.3%", result.Usage.MemoryUsagePercent)

	assert.Empty(t, writer.String())
}

func TestRunWithMetricsUnavailable(t *testing.T) {
	// Given
	var writer bytes.Buffer
	node := createTestNode("node-metrics-test", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0")

	fakeKubeClient := kubefake.NewSimpleClientset()
	mockClient := &fake.KubeClient{
		TestKubernetesInterface: fakeKubeClient,
	}

	_, err := mockClient.TestKubernetesInterface.CoreV1().Nodes().Create(context.TODO(), &node, metav1.CreateOptions{})
	assert.NoError(t, err)

	collector := diagnostics.NewNodeResourceInfoCollector(mockClient, &writer, true)

	// When
	results := collector.Run(context.TODO())

	// Then
	assert.Len(t, results, 1)
	result := results[0]

	assertNodeBasicInfo(t, result, node)

	// Verify new metrics fields are present
	assert.NotNil(t, result.Usage.CPUUsage)
	assert.NotNil(t, result.Usage.MemoryUsage)
	assert.NotNil(t, result.Usage.CPUUsagePercent)
	assert.NotNil(t, result.Usage.MemoryUsagePercent)
	assert.NotNil(t, result.Usage.CPULimitPercent)
	assert.NotNil(t, result.Usage.MemoryLimitPercent)

	// Since metrics server is not available, these should be "N/A"
	assert.Equal(t, "N/A", result.Usage.CPUUsage)
	assert.Equal(t, "N/A", result.Usage.MemoryUsage)
	assert.Equal(t, "N/A", result.Usage.CPUUsagePercent)
	assert.Equal(t, "N/A", result.Usage.MemoryUsagePercent)
}

// createTestNode creates a test node with the specified properties
func createTestNode(name, arch, kernelVersion, osImage, containerRuntime, kubeletVersion string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"topology.kubernetes.io/region": "region",
				"topology.kubernetes.io/zone":   "zone",
			},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("4"),
				corev1.ResourceMemory:           resource.MustParse("8Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
				corev1.ResourcePods:             resource.MustParse("110"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("3900m"),
				corev1.ResourceMemory:           resource.MustParse("7Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("95Gi"),
				corev1.ResourcePods:             resource.MustParse("110"),
			},
			NodeInfo: corev1.NodeSystemInfo{
				Architecture:            arch,
				KernelVersion:           kernelVersion,
				OSImage:                 osImage,
				ContainerRuntimeVersion: containerRuntime,
				KubeletVersion:          kubeletVersion,
				OperatingSystem:         "linux",
			},
		},
	}
}

func createTestPodWithResources(name, nodeName string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}
