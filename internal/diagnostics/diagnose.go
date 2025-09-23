package diagnostics

import (
	"context"
	"os"

	"github.com/kyma-project/cli.v3/internal/kube"
)

type DiagnosticData struct {
	Metadata      Metadata
	Warnings      KymaSystemWarnings
	NodeResources []NodeResourceInfo
}

func GetData(ctx context.Context, client kube.Client, verbose bool) DiagnosticData {
	metadataCollector := NewMetadataCollector(client, os.Stdout, verbose)
	kymaSystemWarningsCollector := NewKymaSystemWarningsCollector(client, os.Stdout, verbose)
	nodeResourceInfoCollector := NewNodeResourceInfoCollector(client, os.Stdout, verbose)

	return DiagnosticData{
		Metadata:      metadataCollector.Run(ctx),
		Warnings:      kymaSystemWarningsCollector.Run(ctx),
		NodeResources: nodeResourceInfoCollector.Run(ctx),
	}
}
