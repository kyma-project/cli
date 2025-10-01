package diagnostics

import (
	"context"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterWarnings struct {
	Warnings []corev1.Event
}

type ClusterWarningsCollector struct {
	client kube.Client
	VerboseLogger
}

func NewClusterWarningsCollector(client kube.Client, writer io.Writer, verbose bool) *ClusterWarningsCollector {
	return &ClusterWarningsCollector{
		client:        client,
		VerboseLogger: NewVerboseLogger(writer, verbose),
	}
}

func (wc *ClusterWarningsCollector) Run(ctx context.Context) ClusterWarnings {
	warnings, err := wc.getClusterWarnings(ctx)
	if err != nil {
		wc.WriteVerboseError(err, "Failed to get system warnings from the cluster")
	}

	return ClusterWarnings{
		Warnings: warnings,
	}
}

func (wc *ClusterWarningsCollector) getClusterWarnings(ctx context.Context) ([]corev1.Event, error) {
	allEvents, err := wc.getClusterEvents(ctx)
	if err != nil {
		return nil, err
	}

	var warnings []corev1.Event
	for _, event := range allEvents {
		if event.Type == "Warning" {
			warnings = append(warnings, event)
		}
	}

	return warnings, nil
}

func (wc *ClusterWarningsCollector) getClusterEvents(ctx context.Context) ([]corev1.Event, error) {
	eventList, err := wc.client.Static().CoreV1().
		Events("").
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return eventList.Items, nil
}
