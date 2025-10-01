package diagnostics

import (
	"context"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventInfo struct {
	InvolvedObject corev1.ObjectReference `json:"involvedObject" yaml:"involvedObject"`
	Reason         string                 `json:"reason" yaml:"reason"`
	Message        string                 `json:"message" yaml:"message"`
	Count          int32                  `json:"count" yaml:"count"`
	EventTime      metav1.MicroTime       `json:"eventTime" yaml:"eventTime"`
	Namespace      string                 `json:"namespace" yaml:"namespace"`
}

type ClusterWarnings struct {
	Warnings []EventInfo `json:"warnings" yaml:"warnings"`
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

func (wc *ClusterWarningsCollector) getClusterWarnings(ctx context.Context) ([]EventInfo, error) {
	allEvents, err := wc.getClusterEvents(ctx)
	if err != nil {
		return nil, err
	}

	var warnings []EventInfo
	for _, event := range allEvents {
		if event.Type == "Warning" {
			eventInfo := EventInfo{
				InvolvedObject: event.InvolvedObject,
				Reason:         event.Reason,
				Message:        event.Message,
				Count:          event.Count,
				EventTime:      event.EventTime,
				Namespace:      event.Namespace,
			}
			warnings = append(warnings, eventInfo)
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
