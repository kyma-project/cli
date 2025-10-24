package diagnostics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

type NodeMetrics struct {
	Usage struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
	} `json:"usage"`
}

type MachineInfo struct {
	Name             string `json:"name" yaml:"name"`
	Architecture     string `json:"architecture" yaml:"architecture"`
	KernelVersion    string `json:"kernelVersion" yaml:"kernelVersion"`
	OSImage          string `json:"osImage" yaml:"osImage"`
	ContainerRuntime string `json:"containerRuntime" yaml:"containerRuntime"`
	KubeletVersion   string `json:"kubeletVersion" yaml:"kubeletVersion"`
	OperatingSystem  string `json:"operatingSystem" yaml:"operatingSystem"`
}

type Capacity struct {
	CPU              string `json:"cpu" yaml:"cpu"`
	Memory           string `json:"memory" yaml:"memory"`
	EphemeralStorage string `json:"ephemeralStorage" yaml:"ephemeralStorage"`
	Pods             string `json:"pods" yaml:"pods"`
}

// Usage represents comprehensive resource information for the node
type Usage struct {
	// Actual usage data from the metrics server
	CPUUsage           string `json:"cpuUsage" yaml:"cpuUsage"`
	MemoryUsage        string `json:"memoryUsage" yaml:"memoryUsage"`
	CPUUsagePercent    string `json:"cpuUsagePercent" yaml:"cpuUsagePercent"`       // Percentage of allocatable CPU currently being used
	MemoryUsagePercent string `json:"memoryUsagePercent" yaml:"memoryUsagePercent"` // Percentage of allocatable memory currently being used

	// Resource requests
	CPURequests            string `json:"cpuRequests" yaml:"cpuRequests"`                       // Total CPU requests from all pods (e.g., "1200m")
	MemoryRequests         string `json:"memoryRequests" yaml:"memoryRequests"`                 // Total memory requests from all pods (e.g., "2.5Gi")
	CPURequestedPercent    string `json:"cpuRequestedPercent" yaml:"cpuRequestedPercent"`       // Percentage of allocatable CPU requested
	MemoryRequestedPercent string `json:"memoryRequestedPercent" yaml:"memoryRequestedPercent"` // Percentage of allocatable memory requested

	// Resource limits
	CPULimits          string `json:"cpuLimits" yaml:"cpuLimits"`                   // Total CPU limits from all pods (e.g., "2400m")
	MemoryLimits       string `json:"memoryLimits" yaml:"memoryLimits"`             // Total memory limits from all pods (e.g., "4Gi")
	CPULimitPercent    string `json:"cpuLimitPercent" yaml:"cpuLimitPercent"`       // Percentage of allocatable CPU limited
	MemoryLimitPercent string `json:"memoryLimitPercent" yaml:"memoryLimitPercent"` // Percentage of allocatable memory limited

	// Available resources (what's left for new workloads)
	CPUAvailable              string `json:"cpuAvailable" yaml:"cpuAvailable"`                           // Available CPU for new pods (Allocatable - Requests)
	MemoryAvailable           string `json:"memoryAvailable" yaml:"memoryAvailable"`                     // Available memory for new pods (Allocatable - Requests)
	EphemeralStorageAvailable string `json:"ephemeralStorageAvailable" yaml:"ephemeralStorageAvailable"` // Available storage (typically equals allocatable)
	PodsAvailable             string `json:"podsAvailable" yaml:"podsAvailable"`                         // Available pod slots (e.g., "89" if 110 max - 21 running)

	// Pod information
	PodCount int `json:"podCount" yaml:"podCount"` // Number of running/pending pods on this node
}

type Topology struct {
	Region string `json:"region" yaml:"region"`
	Zone   string `json:"zone" yaml:"zone"`
}

type NodeResourceInfo struct {
	MachineInfo MachineInfo `json:"machineInfo" yaml:"machineInfo"`
	Capacity    Capacity    `json:"capacity" yaml:"capacity"`
	Usage       Usage       `json:"usage" yaml:"usage"`
	Topology    Topology    `json:"topology" yaml:"topology"`
}

type NodeResourceInfoCollector struct {
	client kube.Client
	*out.Printer
}

func NewNodeResourceInfoCollector(client kube.Client) *NodeResourceInfoCollector {
	return &NodeResourceInfoCollector{
		client:  client,
		Printer: out.Default,
	}
}

func (c *NodeResourceInfoCollector) Run(ctx context.Context) []NodeResourceInfo {
	nodes, err := c.client.Static().CoreV1().
		Nodes().
		List(ctx, metav1.ListOptions{})

	if err != nil {
		c.Verbosefln("Failed to list nodes from the cluster: %v", err)
		return []NodeResourceInfo{}
	}

	var nodeResources []NodeResourceInfo

	for _, node := range nodes.Items {
		// Get pods running on this node
		pods, err := c.getPodsOnNode(ctx, node.Name)
		if err != nil {
			c.Verbosefln("Failed to get pods for node %s: %v", node.Name, err)
		}

		usage := c.calculateUsage(ctx, node, pods)
		topology := c.getTopologyData(node)

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
			Usage:    usage,
			Topology: topology,
		}

		nodeResources = append(nodeResources, nodeInfo)
	}

	return nodeResources
}

func (c *NodeResourceInfoCollector) getTopologyData(node corev1.Node) Topology {
	return Topology{
		Region: node.ObjectMeta.Labels["topology.kubernetes.io/region"],
		Zone:   node.ObjectMeta.Labels["topology.kubernetes.io/zone"],
	}
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

func (c *NodeResourceInfoCollector) calculateUsage(ctx context.Context, node corev1.Node, pods []corev1.Pod) Usage {
	cpuRequests, memoryRequests, cpuLimits, memoryLimits :=
		c.calculateRequestsAndLimits(pods)
	cpuAvailable, memoryAvailable, ephemeralStorageAvailable, podsAvailable :=
		c.calculateAvailableResources(node, cpuRequests, memoryRequests, len(pods))
	cpuRequestedPercent, memoryRequestedPercent :=
		c.calculateRequestedPercentages(node, cpuRequests, memoryRequests)
	cpuUsage, memoryUsage, cpuUsagePercent, memoryUsagePercent :=
		c.calculateActualUsage(ctx, node.Name, node)
	cpuLimitPercent, memoryLimitPercent :=
		c.calculateLimitPercentages(node, cpuLimits, memoryLimits)

	return Usage{
		// Actual usage data
		CPUUsage:           cpuUsage,
		MemoryUsage:        memoryUsage,
		CPUUsagePercent:    cpuUsagePercent,
		MemoryUsagePercent: memoryUsagePercent,

		// Requests
		CPURequests:            cpuRequests.String(),
		MemoryRequests:         memoryRequests.String(),
		CPURequestedPercent:    cpuRequestedPercent,
		MemoryRequestedPercent: memoryRequestedPercent,

		// Limits
		CPULimits:          cpuLimits.String(),
		MemoryLimits:       memoryLimits.String(),
		CPULimitPercent:    cpuLimitPercent,
		MemoryLimitPercent: memoryLimitPercent,

		// Available resources
		CPUAvailable:              cpuAvailable,
		MemoryAvailable:           memoryAvailable,
		EphemeralStorageAvailable: ephemeralStorageAvailable,
		PodsAvailable:             podsAvailable,

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

func (c *NodeResourceInfoCollector) getNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	restClient := c.client.RestClient()
	if restClient == nil {
		return nil, fmt.Errorf("REST client is not available")
	}

	result := restClient.Get().
		AbsPath("/apis/metrics.k8s.io/v1beta1/nodes/" + nodeName).
		Do(ctx)

	if result.Error() != nil {
		return nil, fmt.Errorf("failed to get metrics for node %s: %w", nodeName, result.Error())
	}

	rawData, err := result.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw response: %w", err)
	}

	var nodeMetrics NodeMetrics
	if err := json.Unmarshal(rawData, &nodeMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics response: %w", err)
	}

	return &nodeMetrics, nil
}

func formatMemoryToMiBFromString(memStr string) string {
	quantity, err := resource.ParseQuantity(memStr)
	if err != nil {
		return memStr
	}
	return formatMemoryToMiB(quantity)
}

func formatMemoryToMiB(quantity resource.Quantity) string {
	bytes := quantity.Value()
	mib := float64(bytes) / (1024 * 1024)
	return fmt.Sprintf("%.1fMi", mib)
}

// convertCPUNanosToMillicores converts CPU nanoseconds string to millicores
// Returns the millicores value and whether the conversion was successful
func convertCPUNanosToMillicores(cpuStr string) (int64, bool) {
	// Remove the 'n' suffix if present
	if len(cpuStr) > 0 && cpuStr[len(cpuStr)-1] == 'n' {
		nanosStr := cpuStr[:len(cpuStr)-1]

		// Parse nanoseconds
		nanos, err := strconv.ParseInt(nanosStr, 10, 64)
		if err != nil {
			return 0, false
		}

		// Convert nanoseconds to millicores
		// 1 CPU core = 1,000,000,000 nanoseconds per second
		// 1 millicore = 1,000,000 nanoseconds per second
		millicores := nanos / 1000000
		return millicores, true
	}

	return 0, false
}

func formatCPUNanos(cpuStr string) string {
	if millicores, ok := convertCPUNanosToMillicores(cpuStr); ok {
		return fmt.Sprintf("%.1fm", float64(millicores))
	}
	return cpuStr
}

func (c *NodeResourceInfoCollector) calculateLimitPercentages(node corev1.Node, cpuLimits, memoryLimits resource.Quantity) (string, string) {
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memoryAllocatable := node.Status.Allocatable.Memory()

	var cpuLimitPercent, memoryLimitPercent string
	if !cpuAllocatable.IsZero() {
		cpuLimitPercent = calculatePercentage(cpuLimits, *cpuAllocatable)
	}
	if !memoryAllocatable.IsZero() {
		memoryLimitPercent = calculatePercentage(memoryLimits, *memoryAllocatable)
	}

	return cpuLimitPercent, memoryLimitPercent
}

func (c *NodeResourceInfoCollector) calculateActualUsage(ctx context.Context, nodeName string, node corev1.Node) (string, string, string, string) {
	metrics, err := c.getNodeMetrics(ctx, nodeName)
	if err != nil {
		c.Verbosefln("Failed to get metrics for node %s: %v", nodeName, err)
		return "N/A", "N/A", "N/A", "N/A"
	}

	cpuUsage := formatCPUNanos(metrics.Usage.CPU)
	memoryUsage := formatMemoryToMiBFromString(metrics.Usage.Memory)

	cpuUsagePercent, memoryUsagePercent := c.calculateUsagePercentages(metrics, node)

	return cpuUsage, memoryUsage, cpuUsagePercent, memoryUsagePercent
}

func (c *NodeResourceInfoCollector) calculateUsagePercentages(metrics *NodeMetrics, node corev1.Node) (string, string) {
	cpuAllocatable := node.Status.Allocatable.Cpu()
	memoryAllocatable := node.Status.Allocatable.Memory()

	var cpuUsagePercent, memoryUsagePercent string

	if !cpuAllocatable.IsZero() {
		if cpuUsage, err := parseActualCPUUsage(metrics.Usage.CPU); err == nil {
			cpuUsagePercent = calculatePercentage(cpuUsage, *cpuAllocatable)
		} else {
			cpuUsagePercent = "N/A"
		}
	}

	if !memoryAllocatable.IsZero() {
		if memoryUsage, err := resource.ParseQuantity(metrics.Usage.Memory); err == nil {
			memoryUsagePercent = calculatePercentage(memoryUsage, *memoryAllocatable)
		} else {
			memoryUsagePercent = "N/A"
		}
	}

	return cpuUsagePercent, memoryUsagePercent
}

func parseActualCPUUsage(cpuStr string) (resource.Quantity, error) {
	if millicores, ok := convertCPUNanosToMillicores(cpuStr); ok {
		return *resource.NewMilliQuantity(millicores, resource.DecimalSI), nil
	}

	// Fall back to standard quantity parsing
	return resource.ParseQuantity(cpuStr)
}
