package logs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// FetcherForTestingPods provides functionality for fetching logs from test suite results
type FetcherForTestingPods struct {
	ignoredContainers map[string]struct{}
	podCli            v1.PodsGetter
}

// NewFetcherForTestingPods returns new instance of the FetcherForTestingPods
func NewFetcherForTestingPods(podCli v1.PodsGetter, ignoredContainers []string) *FetcherForTestingPods {
	f := FetcherForTestingPods{
		ignoredContainers: map[string]struct{}{},
		podCli:            podCli,
	}

	for _, c := range ignoredContainers {
		f.ignoredContainers[c] = struct{}{}
	}

	return &f
}

// Logs returns logs from given test suite results. Respects the ignored containers
// from where the logs are not fetched.
func (f *FetcherForTestingPods) Logs(result oct.TestResult) (string, error) {
	logs := strings.Builder{}
	for _, exec := range result.Executions {
		pod, err := f.podCli.Pods(result.Namespace).Get(context.Background(), exec.ID, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "while getting %q pod", exec.ID)
		}

		var dumpContainers []string
		for _, c := range pod.Spec.Containers {
			if _, skip := f.ignoredContainers[c.Name]; skip {
				continue
			}
			dumpContainers = append(dumpContainers, c.Name)
		}

		for _, c := range dumpContainers {
			out, err := f.podCli.Pods(result.Namespace).GetLogs(exec.ID, &corev1.PodLogOptions{
				Container: c,
				Follow:    false,
			}).DoRaw(context.Background())

			if err != nil {
				return "", errors.Wrapf(err, "while fetching logs for pod %q", exec.ID)
			}

			logs.WriteString(fmt.Sprintf("Start of logs from container %q in pod %q in status %q\n", c, exec.ID, result.Status))
			logs.WriteString(fmt.Sprintf("%s", out))
			logs.WriteString(fmt.Sprintf("End of logs from container %q in pod %q in status %q\n", c, exec.ID, result.Status))

		}
	}

	return strip(logs.String()), nil
}

func strip(s string) string {
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(s, "")
}
