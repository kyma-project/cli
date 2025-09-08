package diagnostics

import (
	"context"
	"fmt"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceMetrics struct {
	CPU              string
	Memory           string
	EphemeralStorage string
	Pods             string
}

type NodeResourceInfo struct {
	Name             string
	Capacity         ResourceMetrics
	Allocatable      ResourceMetrics
	Architecture     string
	KernelVersion    string
	OSImage          string
	ContainerRuntime string
	KubeletVersion   string
}

type NodeResourceInfoCollector struct {
	client  kube.Client
	writer  io.Writer
	verbose bool
}

func NewNodeResourceInfoCollector(client kube.Client, writer io.Writer, verbose bool) *NodeResourceInfoCollector {
	return &NodeResourceInfoCollector{
		client:  client,
		writer:  writer,
		verbose: verbose,
	}
}

func (ric *NodeResourceInfoCollector) Run(ctx context.Context) []NodeResourceInfo {
	nodes, err := ric.client.Static().CoreV1().
		Nodes().
		List(ctx, metav1.ListOptions{})

	if err != nil {
		ric.WriteVerboseError(err, "Failed to list nodes from the cluster")
	}

	var nodeResources []NodeResourceInfo

	for _, node := range nodes.Items {
		nodeInfo := NodeResourceInfo{
			Name: node.Name,
			Capacity: ResourceMetrics{
				CPU:              node.Status.Capacity.Cpu().String(),
				Memory:           node.Status.Capacity.Memory().String(),
				EphemeralStorage: node.Status.Capacity.StorageEphemeral().String(),
				Pods:             node.Status.Capacity.Pods().String(),
			},
			Allocatable: ResourceMetrics{
				CPU:              node.Status.Allocatable.Cpu().String(),
				Memory:           node.Status.Allocatable.Memory().String(),
				EphemeralStorage: node.Status.Allocatable.StorageEphemeral().String(),
				Pods:             node.Status.Allocatable.Pods().String(),
			},
			Architecture:     node.Status.NodeInfo.Architecture,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		}

		nodeResources = append(nodeResources, nodeInfo)
	}

	return nodeResources
}

func (ric *NodeResourceInfoCollector) WriteVerboseError(err error, message string) {
	if !ric.verbose || err == nil {
		return
	}

	fmt.Fprintf(ric.writer, "%s: %s\n", message, err.Error())
}
