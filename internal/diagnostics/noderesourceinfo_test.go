package diagnostics_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

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

func TestNodeResourceInfoWriteVerboseError(t *testing.T) {
	testCases := []struct {
		name           string
		verbose        bool
		err            error
		message        string
		expectedOutput string
	}{
		{
			name:           "Should write error when verbose is true",
			verbose:        true,
			err:            errors.New("test error"),
			message:        "Test error message",
			expectedOutput: "Test error message: test error\n",
		},
		{
			name:           "Should not write error when verbose is false",
			verbose:        false,
			err:            errors.New("test error"),
			message:        "Test error message",
			expectedOutput: "",
		},
		{
			name:           "Should not write error when error is nil",
			verbose:        true,
			err:            nil,
			message:        "Test error message",
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			collector := diagnostics.NewNodeResourceInfoCollector(nil, &writer, tc.verbose)

			// When
			collector.WriteVerboseError(tc.err, tc.message)

			// Then
			assert.Equal(t, tc.expectedOutput, writer.String())
		})
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name            string
		nodes           []corev1.Node
		expectError     bool
		expectedResults int
		verbose         bool
	}{
		{
			name:            "Should return empty list when no nodes exist",
			nodes:           []corev1.Node{},
			expectError:     false,
			expectedResults: 0,
			verbose:         false,
		},
		{
			name: "Should collect node information when single node exists",
			nodes: []corev1.Node{
				createTestNode("node1", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0"),
			},
			expectError:     false,
			expectedResults: 1,
			verbose:         false,
		},
		{
			name: "Should collect node information when multiple nodes exist",
			nodes: []corev1.Node{
				createTestNode("node1", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0"),
				createTestNode("node2", "arm64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.1"),
				createTestNode("node3", "amd64", "5.8.0-generic", "Ubuntu 22.04.1 LTS", "containerd://1.6.2", "v1.26.2"),
			},
			expectError:     false,
			expectedResults: 3,
			verbose:         false,
		},
		{
			name: "Should calculate usage correctly with running pods",
			nodes: []corev1.Node{
				createTestNode("node-with-pods", "amd64", "5.4.0-generic", "Ubuntu 20.04.3 LTS", "containerd://1.5.9", "v1.26.0"),
			},
			expectError:     false,
			expectedResults: 1,
			verbose:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeKubeClient := kubefake.NewSimpleClientset()
			mockClient := &fake.KubeClient{
				TestKubernetesInterface: fakeKubeClient,
			}

			// Create nodes in the fake client
			for _, node := range tc.nodes {
				_, err := fakeKubeClient.CoreV1().Nodes().Create(context.TODO(), &node, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			// Create test pods for the usage calculation test
			if tc.name == "Should calculate usage correctly with running pods" && len(tc.nodes) > 0 {
				testPod := createTestPodWithResources("test-pod", tc.nodes[0].Name)
				_, err := fakeKubeClient.CoreV1().Pods("default").Create(context.TODO(), &testPod, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			collector := diagnostics.NewNodeResourceInfoCollector(mockClient, &writer, tc.verbose)

			// When
			results := collector.Run(context.TODO())

			// Then
			assert.Len(t, results, tc.expectedResults)

			// Verify the collected data matches what we expect
			if len(tc.nodes) > 0 {
				for i, result := range results {
					expectedNode := tc.nodes[i]

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

					// Check merged usage information (includes availability)
					if tc.name == "Should calculate usage correctly with running pods" {
						// Resource totals
						assert.Equal(t, "100m", result.Usage.CPURequests)
						assert.Equal(t, "128Mi", result.Usage.MemoryRequests)
						assert.Equal(t, "200m", result.Usage.CPULimits)
						assert.Equal(t, "256Mi", result.Usage.MemoryLimits)

						// Available resources (allocatable - requests)
						assert.Equal(t, "3800m", result.Usage.CPUAvailable) // 3900m - 100m
						assert.Equal(t, "109", result.Usage.PodsAvailable)  // 110 - 1
						assert.Equal(t, expectedNode.Status.Allocatable.StorageEphemeral().String(), result.Usage.EphemeralStorageAvailable)

						// Percentage calculations
						assert.Equal(t, "2.6%", result.Usage.CPURequestedPercent) // 100m / 3900m * 100

						// Pod count
						assert.Equal(t, 1, result.Usage.PodCount)
					} else {
						// No pods scenarios
						assert.Equal(t, "0", result.Usage.CPURequests)
						assert.Equal(t, "0", result.Usage.MemoryRequests)
						assert.Equal(t, "0", result.Usage.CPULimits)
						assert.Equal(t, "0", result.Usage.MemoryLimits)

						// Available should equal allocatable when no pods
						assert.Equal(t, expectedNode.Status.Allocatable.Cpu().String(), result.Usage.CPUAvailable)
						assert.Equal(t, expectedNode.Status.Allocatable.Memory().String(), result.Usage.MemoryAvailable)
						assert.Equal(t, expectedNode.Status.Allocatable.Pods().String(), result.Usage.PodsAvailable)

						// Percentages should be 0%
						assert.Equal(t, "0.0%", result.Usage.CPURequestedPercent)
						assert.Equal(t, "0.0%", result.Usage.MemoryRequestedPercent)

						// Pod count should be 0
						assert.Equal(t, 0, result.Usage.PodCount)
					}
				}
			}

			// Check if verbose error logging works as expected
			if !tc.expectError {
				assert.Empty(t, writer.String())
			}
		})
	}
}

// createTestNode creates a test node with the specified properties
func createTestNode(name, arch, kernelVersion, osImage, containerRuntime, kubeletVersion string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
