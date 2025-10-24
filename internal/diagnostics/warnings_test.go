package diagnostics

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/out"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewClusterWarningsCollector(t *testing.T) {
	// Given
	fakeClient := &fake.KubeClient{}

	// When
	collector := NewClusterWarningsCollector(fakeClient)

	// Then
	assert.NotNil(t, collector)
}

func TestClusterWarningsCollector_Run(t *testing.T) {
	testCases := []struct {
		name                  string
		mockEvents            []corev1.Event
		verbose               bool
		expectedWarningsCount int
	}{
		{
			name: "Should collect warnings successfully",
			mockEvents: []corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "warning-event-1",
						Namespace: "kyma-system",
					},
					Type:           "Warning",
					Reason:         "FailedScheduling",
					Message:        "0/3 nodes are available",
					FirstTimestamp: metav1.NewTime(time.Now()),
					LastTimestamp:  metav1.NewTime(time.Now()),
					Count:          1,
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "normal-event",
						Namespace: "kyma-system",
					},
					Type:           "Normal",
					Reason:         "Scheduled",
					Message:        "Successfully assigned pod to node",
					FirstTimestamp: metav1.NewTime(time.Now()),
					LastTimestamp:  metav1.NewTime(time.Now()),
					Count:          1,
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "warning-event-2",
						Namespace: "kyma-system",
					},
					Type:           "Warning",
					Reason:         "Unhealthy",
					Message:        "Readiness probe failed",
					FirstTimestamp: metav1.NewTime(time.Now()),
					LastTimestamp:  metav1.NewTime(time.Now()),
					Count:          3,
				},
			},
			verbose:               false,
			expectedWarningsCount: 2,
		},
		{
			name: "Should handle no warning events",
			mockEvents: []corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "normal-event",
						Namespace: "kyma-system",
					},
					Type:    "Normal",
					Reason:  "Started",
					Message: "Container started successfully",
				},
			},
			verbose:               false,
			expectedWarningsCount: 0,
		},
		{
			name:                  "Should handle empty events list",
			mockEvents:            []corev1.Event{},
			verbose:               false,
			expectedWarningsCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer

			// Create fake Kubernetes client with test events
			fakeK8sClient := k8sfake.NewSimpleClientset()
			if len(tc.mockEvents) > 0 {
				for _, event := range tc.mockEvents {
					_, err := fakeK8sClient.CoreV1().Events("kyma-system").Create(
						context.Background(), &event, metav1.CreateOptions{})
					require.NoError(t, err)
				}
			}

			fakeClient := &fake.KubeClient{
				TestKubernetesInterface: fakeK8sClient,
			}

			printer := out.NewToWriter(&writer)
			if tc.verbose {
				printer.EnableVerbose()
			}

			collector := ClusterWarningsCollector{fakeClient, printer}

			// When
			result := collector.Run(context.Background())

			// Then
			assert.Equal(t, tc.expectedWarningsCount, len(result))
		})
	}
}
