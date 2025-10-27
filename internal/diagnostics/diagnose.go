package diagnostics

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
)

type DiagnosticData struct {
	Metadata                  Metadata                    `json:"metadata" yaml:"metadata"`
	Warnings                  []EventInfo                 `json:"warnings" yaml:"warnings"`
	NodeResources             []NodeResourceInfo          `json:"nodes" yaml:"nodes"`
	ModuleCustomResourceState []ModuleCustomResourceState `json:"kymaModulesErrors" yaml:"kymaModulesErrors"`
}

func GetData(ctx context.Context, client kube.Client) DiagnosticData {
	metadataCollector := NewMetadataCollector(client)
	kymaSystemWarningsCollector := NewClusterWarningsCollector(client)
	nodeResourceInfoCollector := NewNodeResourceInfoCollector(client)
	modulesCustomResourceStates := NewModuleCustomResourceStateCollector(client)

	return DiagnosticData{
		Metadata:                  metadataCollector.Run(ctx),
		Warnings:                  kymaSystemWarningsCollector.Run(ctx),
		NodeResources:             nodeResourceInfoCollector.Run(ctx),
		ModuleCustomResourceState: modulesCustomResourceStates.Run(ctx),
	}
}
