package diagnostics

import (
	"context"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
)

type DiagnosticData struct {
	Metadata      Metadata           `json:"metadata" yaml:"metadata"`
	Warnings      KymaSystemWarnings `json:"warnings" yaml:"warnings"`
	NodeResources []NodeResourceInfo `json:"nodes" yaml:"nodes"`
}

func GetData(ctx context.Context, client kube.Client, output io.Writer, verbose bool) DiagnosticData {
	metadataCollector := NewMetadataCollector(client, output, verbose)
	kymaSystemWarningsCollector := NewKymaSystemWarningsCollector(client, output, verbose)
	nodeResourceInfoCollector := NewNodeResourceInfoCollector(client, output, verbose)

	return DiagnosticData{
		Metadata:      metadataCollector.Run(ctx),
		Warnings:      kymaSystemWarningsCollector.Run(ctx),
		NodeResources: nodeResourceInfoCollector.Run(ctx),
	}
}
