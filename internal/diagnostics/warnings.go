package diagnostics

import (
	"context"
	"fmt"
	"io"

	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KymaSystemWarnings struct {
	Warnings []corev1.Event
}

type KymaSystemWarningsCollector struct {
	client  kube.Client
	writer  io.Writer
	verbose bool
}

func NewKymaSystemWarningsCollector(client kube.Client, writer io.Writer, verbose bool) *KymaSystemWarningsCollector {
	return &KymaSystemWarningsCollector{
		client:  client,
		writer:  writer,
		verbose: verbose,
	}
}

func (wc *KymaSystemWarningsCollector) Run(ctx context.Context) KymaSystemWarnings {
	warnings, err := wc.getKymaSystemWarnings(ctx)
	if err != nil {

	}

	return KymaSystemWarnings{
		Warnings: warnings,
	}
}

func (wc *KymaSystemWarningsCollector) getKymaSystemWarnings(ctx context.Context) ([]corev1.Event, error) {
	allEvents, err := wc.getKymaSystemEvents(ctx)
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

func (wc *KymaSystemWarningsCollector) getKymaSystemEvents(ctx context.Context) ([]corev1.Event, error) {
	eventList, err := wc.client.Static().CoreV1().
		Events("kyma-system").
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return eventList.Items, nil
}

func (wc *KymaSystemWarningsCollector) WriteVerboseError(err error, message string) {
	if !wc.verbose || err == nil {
		return
	}

	fmt.Fprintf(wc.writer, "%s: %s\n", message, err.Error())
}
