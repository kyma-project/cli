package diagnostics

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
	Logs   map[string][]string `json:"logs" yaml:"logs"`
	Errors map[string][]string `json:"errors" yaml:"errors"`
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
	errs := make(map[string][]string)

	for _, module := range c.modules {
		moduleLogs, moduleErrs := c.collectModuleLogs(ctx, module)
		// merge moduleLogs into result
		for k, v := range moduleLogs {
			logs[k] = append(logs[k], v...)
		}
		for k, v := range moduleErrs {
			errs[k] = append(errs[k], v...)
		}
	}
	return ModuleLogs{Logs: logs, Errors: errs}
}

// collectModuleLogs collects error logs from a single module and returns its map
func (c *ModuleLogsCollector) collectModuleLogs(ctx context.Context, moduleName string) (map[string][]string, map[string][]string) {
	collected := make(map[string][]string)
	errs := make(map[string][]string)
	pods, err := c.getPodsForModule(ctx, moduleName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errs[moduleName] = append(errs[moduleName], "logs collection cancelled due to timeout")
		} else {
			errs[moduleName] = append(errs[moduleName], fmt.Sprintf("failed to get pods: %v", err))
		}
		return collected, errs
	}

	for i := range pods {
		podLogs, podErrs := c.collectPodLogs(ctx, &pods[i])
		for k, v := range podLogs {
			collected[k] = append(collected[k], v...)
		}
		for k, v := range podErrs {
			errs[k] = append(errs[k], v...)
		}
	}

	return collected, errs
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

func (c *ModuleLogsCollector) collectPodLogs(ctx context.Context, pod *corev1.Pod) (map[string][]string, map[string][]string) {
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	collected := make(map[string][]string)
	errors := make(map[string][]string)

	for _, container := range containers {
		containerLogs, err := c.collectContainerLogs(ctx, pod, container.Name)
		key := fmt.Sprintf("%s/%s", pod.Name, container.Name)
		if err != nil {
			errors[key] = append(errors[key], err.Error())
		}
		collected[key] = append(collected[key], containerLogs...)
	}

	return collected, errors
}

func (c *ModuleLogsCollector) collectContainerLogs(ctx context.Context, pod *corev1.Pod, containerName string) ([]string, error) {
	logOptions := c.buildPodLogOptions(containerName)

	req := c.client.Static().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	logStream, err := req.Stream(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return []string{}, errors.New("logs collection cancelled due to timeout")

		}
		return []string{}, fmt.Errorf("failed to get log stream: %v", err)
	}
	defer logStream.Close()

	return c.extractStructuredOrErrorLogs(logStream)
}

// extractStructuredOrErrorLogs parses each log line strictly as JSON and only keeps lines
// that contain the keys: level, msg, and ts. Non-JSON lines or JSON without these keys are ignored.
// Only entries whose level is one of error, fatal, panic are returned.
func (c *ModuleLogsCollector) extractStructuredOrErrorLogs(logStream io.ReadCloser) ([]string, error) {
	scanner := bufio.NewScanner(logStream)
	var collected []string

	shouldIncludeLine := c.getLogsFilteringStrategy()

	for scanner.Scan() {
		line := scanner.Text()
		if shouldIncludeLine(line) {
			collected = append(collected, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return collected, fmt.Errorf("error reading logs %v", err)
	}

	return collected, nil
}

func (c *ModuleLogsCollector) getLogsFilteringStrategy() func(string) bool {
	if c.strict {
		return parseLogsStrict
	}
	return parseLogsDefault
}

func (c *ModuleLogsCollector) buildPodLogOptions(containerName string) *corev1.PodLogOptions {
	clone := *c.podLogTemplate
	clone.Container = containerName
	return &clone
}

var errorKeywords = []string{
	"error", "exception", "fatal", "panic", "fail", "failed", "failure", "warning", "warn",
}

func parseLogsStrict(line string) bool {
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

func parseLogsDefault(line string) bool {
	lineLower := strings.ToLower(line)

	if onlyContainsFalsePositives(lineLower) {
		return false
	}

	for _, keyword := range errorKeywords {
		if strings.Contains(lineLower, keyword) {
			return true
		}
	}

	return false
}

func onlyContainsFalsePositives(line string) bool {
	falsePositives := []string{
		"\"error\": null",
		"\"error\":null",
	}

	hasFalsePositive := false
	for _, fp := range falsePositives {
		if strings.Contains(line, fp) {
			hasFalsePositive = true
			break
		}
	}

	if !hasFalsePositive {
		return false
	}

	cleanedLine := line
	for _, fp := range falsePositives {
		cleanedLine = strings.ReplaceAll(cleanedLine, fp, "")
	}
	cleanedLineLower := strings.ToLower(cleanedLine)

	for _, keyword := range errorKeywords {
		if strings.Contains(cleanedLineLower, keyword) {
			return false
		}
	}

	return true
}
