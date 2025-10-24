package diagnostics

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventInfo struct {
	InvolvedObject corev1.ObjectReference `json:"involvedObject" yaml:"involvedObject"`
	Reason         string                 `json:"reason" yaml:"reason"`
	Message        string                 `json:"message" yaml:"message"`
	Count          int32                  `json:"count" yaml:"count"`
	Source         corev1.EventSource     `json:"source" yaml:"source"`
	EventTime      metav1.MicroTime       `json:"eventTime" yaml:"eventTime"`
	Namespace      string                 `json:"namespace" yaml:"namespace"`
}

type ClusterWarningsCollector struct {
	client kube.Client
	*out.Printer
}

func NewClusterWarningsCollector(client kube.Client) *ClusterWarningsCollector {
	return &ClusterWarningsCollector{
		client:  client,
		Printer: out.Default,
	}
}

func (wc *ClusterWarningsCollector) Run(ctx context.Context) []EventInfo {
	warnings, err := wc.getClusterWarnings(ctx)
	if err != nil {
		wc.Verbosefln("Failed to get system warnings from the cluster: %v", err)
	}

	return warnings
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
				Source:         event.Source,
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
