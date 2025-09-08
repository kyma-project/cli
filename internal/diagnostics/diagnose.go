package diagnostics

import (
	"context"
	"os"

	"github.com/kyma-project/cli.v3/internal/kube"
)

type DiagnosticData struct {
	Metadata       Metadata
	Warnings       KymaSystemWarnings
	NodeResources  []NodeResourceInfo
	ModuleStatuses []ModuleCustomResourceState
}

func GetData(ctx context.Context, client kube.Client) DiagnosticData {
	metadataCollector := NewMetadataCollector(client, os.Stdin, true)
	kymaSystemWarningsCollector := NewKymaSystemWarningsCollector(client, os.Stdin, true)
	nodeResourceInfoCollector := NewNodeResourceInfoCollector(client, os.Stdin, true)
	moduleCustomResourceStateCollector := NewModuleCustomResourceStateCollector(client, os.Stdin, true)

	return DiagnosticData{
		Metadata:       metadataCollector.Run(ctx),
		Warnings:       kymaSystemWarningsCollector.Run(ctx),
		NodeResources:  nodeResourceInfoCollector.Run(ctx),
		ModuleStatuses: moduleCustomResourceStateCollector.Run(ctx),
	}
}
