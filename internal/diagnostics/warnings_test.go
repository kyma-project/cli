package diagnostics_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewClusterWarningsCollector(t *testing.T) {
	// Given
	fakeClient := &fake.KubeClient{}
	var writer bytes.Buffer
	verbose := true

	// When
	collector := diagnostics.NewClusterWarningsCollector(fakeClient, &writer, verbose)

	// Then
	assert.NotNil(t, collector)
}

func TestClusterWarningsCollector_WriteVerboseError(t *testing.T) {
	testCases := []struct {
		name           string
		verbose        bool
		err            error
		message        string
		expectedOutput string
	}{
		{
			name:           "Should write error when verbose is true",
			verbose:        true,
			err:            fmt.Errorf("test error"),
			message:        "Test error message",
			expectedOutput: "Test error message: test error\n",
		},
		{
			name:           "Should not write error when verbose is false",
			verbose:        false,
			err:            fmt.Errorf("test error"),
			message:        "Test error message",
			expectedOutput: "",
		},
		{
			name:           "Should not write error when error is nil",
			verbose:        true,
			err:            nil,
			message:        "Test error message",
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			collector := diagnostics.NewClusterWarningsCollector(nil, &writer, tc.verbose)

			// When
			collector.WriteVerboseError(tc.err, tc.message)

			// Then
			assert.Equal(t, tc.expectedOutput, writer.String())
		})
	}
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

			collector := diagnostics.NewClusterWarningsCollector(fakeClient, &writer, tc.verbose)

			// When
			result := collector.Run(context.Background())

			// Then
			assert.Equal(t, tc.expectedWarningsCount, len(result.Warnings))
		})
	}
}
