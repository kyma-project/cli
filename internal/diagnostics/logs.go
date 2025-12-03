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
	Since  time.Duration
	Lines  int64
	Strict bool
}

type ModuleLogsCollector struct {
	client kube.Client
	*out.Printer
	modules        []string
	podLogTemplate *corev1.PodLogOptions // immutable template for each container
	strict         bool
}

func NewModuleLogsCollector(client kube.Client, modules []string, options LogOptions) *ModuleLogsCollector {
	// precedence: Since > Lines
	podOpts := &corev1.PodLogOptions{}
	if options.Since > 0 {
		since := metav1.NewTime(time.Now().Add(-options.Since))
		podOpts.SinceTime = &since
	} else if options.Lines > 0 {
		lines := options.Lines
		podOpts.TailLines = &lines
	}
	return &ModuleLogsCollector{
		client:         client,
		Printer:        out.Default,
		modules:        modules,
		podLogTemplate: podOpts,
		strict:         options.Strict,
	}
}

func (c *ModuleLogsCollector) Run(ctx context.Context) ModuleLogs {
	logs := make(map[string][]string)

	for _, module := range c.modules {
		moduleLogs := c.collectModuleLogs(ctx, module)
		// merge moduleLogs into result
		for k, v := range moduleLogs {
			logs[k] = append(logs[k], v...)
		}
	}
	return ModuleLogs{Logs: logs}
}

// collectModuleLogs collects error logs from a single module and returns its map
func (c *ModuleLogsCollector) collectModuleLogs(ctx context.Context, moduleName string) map[string][]string {
	collected := make(map[string][]string)
	pods, err := c.getPodsForModule(ctx, moduleName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			out.Errfln("Logs collection for module %s cancelled due to timeout", moduleName)
		} else {
			out.Errfln("Failed to get pods for module %s: %v", moduleName, err)
		}
		return collected
	}

	for i := range pods {
		podLogs := c.collectPodLogs(ctx, &pods[i])
		for k, v := range podLogs {
			collected[k] = append(collected[k], v...)
		}
	}

	return collected
}

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

func (c *ModuleLogsCollector) collectPodLogs(ctx context.Context, pod *corev1.Pod) map[string][]string {
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	collected := make(map[string][]string)

	for _, container := range containers {
		containerLogs := c.collectContainerLogs(ctx, pod, container.Name)
		key := fmt.Sprintf("%s/%s", pod.Name, container.Name)
		collected[key] = append(collected[key], containerLogs...)
	}

	return collected
}

func (c *ModuleLogsCollector) collectContainerLogs(ctx context.Context, pod *corev1.Pod, containerName string) []string {
	logOptions := c.buildPodLogOptions(containerName)

	req := c.client.Static().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	logStream, err := req.Stream(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			out.Errfln("Logs collection cancelled due to timeout for target: %v", fmt.Sprintf("%s/%s", pod.Name, containerName))
		}
		out.Errfln("Failed to get log stream for container %s in pod %s: %v", containerName, pod.Name, err)
		return []string{}
	}
	defer logStream.Close()

	return c.extractStructuredOrErrorLogs(logStream, pod.Name, containerName)
}

// extractStructuredOrErrorLogs parses each log line strictly as JSON and only keeps lines
// that contain the keys: level, msg, and ts. Non-JSON lines or JSON without these keys are ignored.
// Only entries whose level is one of error, fatal, panic are returned.
func (c *ModuleLogsCollector) extractStructuredOrErrorLogs(logStream io.ReadCloser, podName, containerName string) []string {
	scanner := bufio.NewScanner(logStream)
	var collected []string

	isErrorOrWarning := c.getParseStrategy()

	for scanner.Scan() {
		line := scanner.Text()
		if isErrorOrWarning(line) {
			collected = append(collected, line)
		}
	}

	if err := scanner.Err(); err != nil {
		out.Errfln("Error reading logs from pod %s container %s: %v", podName, containerName, err)
	}

	return collected
}

func (c *ModuleLogsCollector) getParseStrategy() func(string) bool {
	if c.strict {
		return c.parseLogsStrict
	}
	return c.parseLogsDefault
}

func (c *ModuleLogsCollector) buildPodLogOptions(containerName string) *corev1.PodLogOptions {
	clone := *c.podLogTemplate
	clone.Container = containerName
	return &clone
}

func (c *ModuleLogsCollector) parseLogsStrict(line string) bool {
	var obj map[string]any

	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		// Not JSON -> ignore
		return false
	}
	lvlRaw, hasLevel := obj["level"]
	_, hasMessage := obj["message"]
	_, hasTimestamp := obj["timestamp"]
	if !hasLevel || (!hasMessage && !hasTimestamp) {
		return false
	}
	lvl := fmt.Sprintf("%v", lvlRaw)
	if !strings.EqualFold(lvl, "error") && !strings.EqualFold(lvl, "warn") && !strings.EqualFold(lvl, "warning") {
		return false
	}

	return true
}

func (c *ModuleLogsCollector) parseLogsDefault(line string) bool {
	lineLower := strings.ToLower(line)

	errorKeywords := []string{
		"error", "exception", "fatal", "panic", "fail", "failed", "failure", "warning", "warn",
	}

	for _, keyword := range errorKeywords {
		if strings.Contains(lineLower, keyword) && !containsFalsePositives(lineLower) {
			return true
		}
	}

	return false
}

func containsFalsePositives(line string) bool {
	falsePositives := []string{
		"\"error\": null", "\"error\":null",
	}

	for _, fp := range falsePositives {
		if strings.Contains(line, fp) {
			return true
		}
	}

	return false
}
