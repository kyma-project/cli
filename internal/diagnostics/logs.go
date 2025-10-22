package diagnostics

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
)

type ModuleLogs struct {
	Logs []string
}

// LogOptions defines the configuration for log collection
type LogOptions struct {
	Since time.Duration
	Lines int64
}

type ModuleLogsCollector struct {
	client kube.Client
	VerboseLogger
}

func NewModuleLogsCollector(client kube.Client, writer io.Writer, verbose bool) *ModuleLogsCollector {
	return &ModuleLogsCollector{
		client:        client,
		VerboseLogger: NewVerboseLogger(writer, verbose),
	}
}

// RunSince collects error logs from specified modules for a given time duration
func (c *ModuleLogsCollector) RunSince(ctx context.Context, modules []string, since time.Duration) ModuleLogs {
	logOptions := LogOptions{Since: since}
	return c.collectLogs(ctx, modules, logOptions)
}

// RunLast collects the last N lines of error logs from specified modules
func (c *ModuleLogsCollector) RunLast(ctx context.Context, modules []string, last int64) ModuleLogs {
	logOptions := LogOptions{Lines: last}
	return c.collectLogs(ctx, modules, logOptions)
}

// collectLogs is the generic method that handles both since and last scenarios
func (c *ModuleLogsCollector) collectLogs(ctx context.Context, modules []string, options LogOptions) ModuleLogs {
	var allLogs []string

	for _, module := range modules {
		moduleLogs := c.collectModuleLogs(ctx, module, options)
		allLogs = append(allLogs, moduleLogs...)
	}

	return ModuleLogs{Logs: allLogs}
}

// collectModuleLogs collects error logs from a single module
func (c *ModuleLogsCollector) collectModuleLogs(ctx context.Context, moduleName string, options LogOptions) []string {
	pods, err := c.getPodsForModule(ctx, moduleName)
	if err != nil {
		c.WriteVerboseError(err, fmt.Sprintf("Failed to get pods for module %s", moduleName))
		return []string{}
	}

	var logs []string
	for _, pod := range pods {
		podLogs := c.collectPodLogs(ctx, &pod, options)
		logs = append(logs, podLogs...)
	}

	return logs
}

// getPodsForModule gets all pods for a specific module in kyma-system namespace
func (c *ModuleLogsCollector) getPodsForModule(ctx context.Context, moduleName string) ([]corev1.Pod, error) {
	labelSelector := map[string]string{
		"kyma-project.io/module": moduleName,
	}

	podList, err := c.client.Static().CoreV1().Pods("kyma-system").List(ctx, metav1.ListOptions{
		LabelSelector: resources.LabelSelectorFor(labelSelector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for module %s: %w", moduleName, err)
	}

	return podList.Items, nil
}

// collectPodLogs collects error logs from a pod
func (c *ModuleLogsCollector) collectPodLogs(ctx context.Context, pod *corev1.Pod, options LogOptions) []string {
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	var logs []string

	for _, container := range containers {
		containerLogs := c.collectContainerLogs(ctx, pod, container.Name, options)
		logs = append(logs, containerLogs...)
	}

	return logs
}

// collectContainerLogs collects and filters error logs from a container
func (c *ModuleLogsCollector) collectContainerLogs(ctx context.Context, pod *corev1.Pod, containerName string, options LogOptions) []string {
	logOptions := c.buildPodLogOptions(containerName, options)

	req := c.client.Static().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	logStream, err := req.Stream(ctx)
	if err != nil {
		c.WriteVerboseError(err, fmt.Sprintf("Failed to get log stream for container %s in pod %s", containerName, pod.Name))
		return []string{}
	}
	defer logStream.Close()

	return c.filterErrorLogs(logStream, pod.Name, containerName)
}

// buildPodLogOptions creates PodLogOptions based on the LogOptions configuration
func (c *ModuleLogsCollector) buildPodLogOptions(containerName string, options LogOptions) *corev1.PodLogOptions {
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
	}

	// Configure based on options - Since takes precedence over Lines
	if options.Since > 0 {
		sinceTime := metav1.NewTime(time.Now().Add(-options.Since))
		logOptions.SinceTime = &sinceTime
	} else if options.Lines > 0 {
		lines := options.Lines
		logOptions.TailLines = &lines
	}

	return logOptions
}

// filterErrorLogs filters log lines for error content and returns them as strings
func (c *ModuleLogsCollector) filterErrorLogs(logStream io.ReadCloser, podName, containerName string) []string {
	scanner := bufio.NewScanner(logStream)
	var errorLogs []string

	// Error tokens to look for (case-insensitive)
	errorTokens := []string{"error", "level=error"}

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		// Check if line contains any error tokens
		hasError := false
		for _, token := range errorTokens {
			if strings.Contains(lineLower, token) {
				hasError = true
				break
			}
		}

		if hasError {
			formattedLog := fmt.Sprintf("[%s/%s] %s", podName, containerName, line)
			errorLogs = append(errorLogs, formattedLog)
		}
	}

	if err := scanner.Err(); err != nil {
		c.WriteVerboseError(err, fmt.Sprintf("Error reading logs from pod %s container %s", podName, containerName))
	}

	return errorLogs
}
