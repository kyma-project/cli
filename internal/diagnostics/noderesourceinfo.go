package diagnostics

import (
	"context"
	"fmt"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

type MachineInfo struct {
	Name             string
	Architecture     string
	KernelVersion    string
	OSImage          string
	ContainerRuntime string
	KubeletVersion   string
	OperatingSystem  string
}

type Capacity struct {
	CPU              string
	Memory           string
	EphemeralStorage string
	Pods             string
}

// Usage represents comprehensive resource information for the node
type Usage struct {
	// Raw resource totals
	CPURequests    string // Total CPU requests from all pods (e.g., "1200m")
	CPULimits      string // Total CPU limits from all pods (e.g., "2400m")
	MemoryRequests string // Total memory requests from all pods (e.g., "2.5Gi")
	MemoryLimits   string // Total memory limits from all pods (e.g., "4Gi")

	// Available resources (what's left for new workloads)
	CPUAvailable              string // Available CPU for new pods (Allocatable - Requests)
	MemoryAvailable           string // Available memory for new pods (Allocatable - Requests)
	EphemeralStorageAvailable string // Available storage (typically equals allocatable)
	PodsAvailable             string // Available pod slots (e.g., "89" if 110 max - 21 running)

	// Usage percentages (how much of allocatable capacity is used)
	CPURequestedPercent    string // Percentage of allocatable CPU requested
	MemoryRequestedPercent string // Percentage of allocatable memory requested

	// Pod information
	PodCount int // Number of running/pending pods on this node
}

type NodeResourceInfo struct {
	MachineInfo MachineInfo
	Capacity    Capacity
	Usage       Usage
}

type NodeResourceInfoCollector struct {
	client kube.Client
	VerboseLogger
}

func NewNodeResourceInfoCollector(client kube.Client, writer io.Writer, verbose bool) *NodeResourceInfoCollector {
	return &NodeResourceInfoCollector{
		client:        client,
		VerboseLogger: NewVerboseLogger(writer, verbose),
	}
}

func (c *NodeResourceInfoCollector) Run(ctx context.Context) []NodeResourceInfo {
	nodes, err := c.client.Static().CoreV1().
		Nodes().
		List(ctx, metav1.ListOptions{})

	if err != nil {
		c.WriteVerboseError(err, "Failed to list nodes from the cluster")
		return []NodeResourceInfo{}
	}

	var nodeResources []NodeResourceInfo

	for _, node := range nodes.Items {
		// Get pods running on this node
		pods, err := c.getPodsOnNode(ctx, node.Name)
		if err != nil {
			c.WriteVerboseError(err, "Failed to get pods for node "+node.Name)
		}

		usage := c.calculateUsage(node, pods)

		nodeInfo := NodeResourceInfo{
			MachineInfo: MachineInfo{
				Name:             node.Name,
				Architecture:     node.Status.NodeInfo.Architecture,
				KernelVersion:    node.Status.NodeInfo.KernelVersion,
				OSImage:          node.Status.NodeInfo.OSImage,
				ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
				KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
				OperatingSystem:  node.Status.NodeInfo.OperatingSystem,
			},
			Capacity: Capacity{
				CPU:              node.Status.Capacity.Cpu().String(),
				Memory:           node.Status.Capacity.Memory().String(),
				EphemeralStorage: node.Status.Capacity.StorageEphemeral().String(),
				Pods:             node.Status.Capacity.Pods().String(),
			},
			Usage: usage,
		}

		nodeResources = append(nodeResources, nodeInfo)
	}

	return nodeResources
}

func (c *NodeResourceInfoCollector) getPodsOnNode(ctx context.Context, nodeName string) ([]corev1.Pod, error) {
	fieldSelector := fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName})
	pods, err := c.client.Static().CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		})

	if err != nil {
		return nil, err
	}

	// Filter out completed/failed pods for accurate usage calculation
	var runningPods []corev1.Pod
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			runningPods = append(runningPods, pod)
		}
	}

	return runningPods, nil
}

func (c *NodeResourceInfoCollector) calculateUsage(node corev1.Node, pods []corev1.Pod) Usage {
	cpuRequests, memoryRequests, cpuLimits, memoryLimits :=
		c.calculateRequestsAndLimits(pods)
	cpuAvailable, memoryAvailable, ephemeralStorageAvailable, podsAvailable :=
		c.calculateAvailableResources(node, cpuRequests, memoryRequests, len(pods))
	cpuRequestedPercent, memoryRequestedPercent :=
		c.calculateRequestedPercentages(node, cpuRequests, memoryRequests)

	return Usage{
		// Raw totals
		CPURequests:    cpuRequests.String(),
		CPULimits:      cpuLimits.String(),
		MemoryRequests: memoryRequests.String(),
		MemoryLimits:   memoryLimits.String(),

		// Available resources
		CPUAvailable:              cpuAvailable,
		MemoryAvailable:           memoryAvailable,
		EphemeralStorageAvailable: ephemeralStorageAvailable,
		PodsAvailable:             podsAvailable,

		// Percentages
		CPURequestedPercent:    cpuRequestedPercent,
		MemoryRequestedPercent: memoryRequestedPercent,

		// Pod count
		PodCount: len(pods),
	}
}

// calculateRequestsAndLimits sums up resource requests and limits from all pods
func (c *NodeResourceInfoCollector) calculateRequestsAndLimits(pods []corev1.Pod) (resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity) {
	var cpuRequests, memoryRequests, cpuLimits, memoryLimits resource.Quantity

	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if req := container.Resources.Requests; req != nil {
				if cpu := req[corev1.ResourceCPU]; !cpu.IsZero() {
					cpuRequests.Add(cpu)
				}
				if memory := req[corev1.ResourceMemory]; !memory.IsZero() {
					memoryRequests.Add(memory)
				}
			}
			if limits := container.Resources.Limits; limits != nil {
				if cpu := limits[corev1.ResourceCPU]; !cpu.IsZero() {
					cpuLimits.Add(cpu)
				}
				if memory := limits[corev1.ResourceMemory]; !memory.IsZero() {
					memoryLimits.Add(memory)
				}
			}
		}
	}

	return cpuRequests, memoryRequests, cpuLimits, memoryLimits
}

// calculateAvailableResources calculates remaining resources available for new workloads
func (c *NodeResourceInfoCollector) calculateAvailableResources(node corev1.Node, cpuRequests, memoryRequests resource.Quantity, podCount int) (string, string, string, string) {
	// Get allocatable resources
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memoryAllocatable := node.Status.Allocatable.Memory()
	storageAllocatable := node.Status.Allocatable.StorageEphemeral()
	podsAllocatable := node.Status.Allocatable.Pods()

	// Calculate available resources (allocatable - requests)
	cpuAvailable := cpuAllocatable.DeepCopy()
	cpuAvailable.Sub(cpuRequests)

	memoryAvailable := memoryAllocatable.DeepCopy()
	memoryAvailable.Sub(memoryRequests)

	podsAvailable := podsAllocatable.DeepCopy()
	podsAvailable.Sub(*resource.NewQuantity(int64(podCount), resource.DecimalSI))

	return cpuAvailable.String(), memoryAvailable.String(), storageAllocatable.String(), podsAvailable.String()
}

// calculateRequestedPercentages calculates what percentage of allocatable resources are requested
func (c *NodeResourceInfoCollector) calculateRequestedPercentages(node corev1.Node, cpuRequests, memoryRequests resource.Quantity) (string, string) {
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memoryAllocatable := node.Status.Allocatable.Memory()

	var cpuRequestedPercent, memoryRequestedPercent string
	if !cpuAllocatable.IsZero() {
		cpuRequestedPercent = calculatePercentage(cpuRequests, *cpuAllocatable)
	}
	if !memoryAllocatable.IsZero() {
		memoryRequestedPercent = calculatePercentage(memoryRequests, *memoryAllocatable)
	}

	return cpuRequestedPercent, memoryRequestedPercent
}

// calculatePercentage calculates the percentage of used vs available resources
func calculatePercentage(used, available resource.Quantity) string {
	if available.IsZero() {
		return "0%"
	}

	usedFloat := float64(used.MilliValue())
	availableFloat := float64(available.MilliValue())

	if availableFloat == 0 {
		return "0%"
	}

	percentage := (usedFloat / availableFloat) * 100
	return fmt.Sprintf("%.1f%%", percentage)
}
