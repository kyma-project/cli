package diagnostics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/out"
)

type ModuleLogs struct {
	Logs map[string][]string
}

type LogOptions struct {
	Since time.Duration
	Lines int64
}

type ModuleLogsCollector struct {
	client kube.Client
	*out.Printer
}

func NewModuleLogsCollector(client kube.Client) *ModuleLogsCollector {
	return &ModuleLogsCollector{
		client:  client,
		Printer: out.Default,
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
	allLogs := make(map[string][]string)

	for _, module := range modules {
		c.collectModuleLogs(ctx, module, options, allLogs)
	}

	return ModuleLogs{Logs: allLogs}
}

// collectModuleLogs collects error logs from a single module
func (c *ModuleLogsCollector) collectModuleLogs(ctx context.Context, moduleName string, options LogOptions, allLogs map[string][]string) {
	pods, err := c.getPodsForModule(ctx, moduleName)
	if err != nil {
		out.Verbosefln("Failed to get pods for module %s: %v", moduleName, err)
	}

	for _, pod := range pods {
		c.collectPodLogs(ctx, &pod, options, allLogs)
	}
}

// getPodsForModule gets all pods for a specific module in kyma-system namespace
func (c *ModuleLogsCollector) getPodsForModule(ctx context.Context, moduleName string) ([]corev1.Pod, error) {
	labelSelector := map[string]string{
		"kyma-project.io/module": moduleName,
	}

	deploymentsList, err := c.client.Static().AppsV1().Deployments("kyma-system").List(ctx, metav1.ListOptions{
		LabelSelector: resources.LabelSelectorFor(labelSelector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments for module %s: %w", moduleName, err)
	}

	podsList := make([]corev1.Pod, 0)

	for _, deployment := range deploymentsList.Items {
		matchLabels := deployment.Spec.Selector.MatchLabels
		pods, err := c.client.Static().CoreV1().Pods("kyma-system").List(ctx, metav1.ListOptions{
			LabelSelector: resources.LabelSelectorFor(matchLabels),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods for deployment %s: %w", deployment.Name, err)
		}

		podsList = append(podsList, pods.Items...)
	}

	return podsList, nil
}

// collectPodLogs collects error logs from a pod
func (c *ModuleLogsCollector) collectPodLogs(ctx context.Context, pod *corev1.Pod, options LogOptions, allLogs map[string][]string) {
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)

	for _, container := range containers {
		containerLogs := c.collectContainerLogs(ctx, pod, container.Name, options)
		key := fmt.Sprintf("%s/%s", pod.Name, container.Name)
		allLogs[key] = append(allLogs[key], containerLogs...)
	}
}

// collectContainerLogs collects and filters error logs from a container
func (c *ModuleLogsCollector) collectContainerLogs(ctx context.Context, pod *corev1.Pod, containerName string, options LogOptions) []string {
	logOptions := c.buildPodLogOptions(containerName, options)

	req := c.client.Static().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	logStream, err := req.Stream(ctx)
	if err != nil {
		out.Verbosefln("Failed to get log stream for container %s in pod %s: %v", containerName, pod.Name, err)
		return []string{}
	}
	defer logStream.Close()

	// Attempt structured JSON parsing; fallback to error token filtering
	return c.extractStructuredOrErrorLogs(logStream, pod.Name, containerName)
}

// extractStructuredOrErrorLogs parses each log line strictly as JSON and only keeps lines
// that contain the keys: level, msg, and ts. Non-JSON lines or JSON without these keys are ignored.
// Only entries whose level is one of error, fatal, panic are returned.
func (c *ModuleLogsCollector) extractStructuredOrErrorLogs(logStream io.ReadCloser, podName, containerName string) []string {
	scanner := bufio.NewScanner(logStream)
	var collected []string

	for scanner.Scan() {
		line := scanner.Text()
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			// Not JSON -> ignore
			continue
		}
		lvlRaw, hasLevel := obj["level"]
		_, hasMsg := obj["msg"]
		_, hasTs := obj["ts"]
		if !hasLevel && !hasMsg && !hasTs {
			continue
		}
		lvl := fmt.Sprintf("%v", lvlRaw)
		if !strings.EqualFold(lvl, "error") && !strings.EqualFold(lvl, "fatal") && !strings.EqualFold(lvl, "panic") {
			continue
		}
		collected = append(collected, line)
	}

	if err := scanner.Err(); err != nil {
		out.Verbosefln("Error reading logs from pod %s container %s: %v", podName, containerName, err)
	}

	return collected
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
