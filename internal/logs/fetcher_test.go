package logs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func Test(t *testing.T) {
	const fixLogsResponse = `Lorem ipsum dolor sit amet.`

	tests := map[string]struct {
		pod        v1.Pod
		ignoredCnt []string
	}{
		"dump all containers": {
			ignoredCnt: nil,
			pod:        fixPodWithContainers("test-1"),
		},
		"skip ignored containers": {
			ignoredCnt: []string{"test-2"},
			pod:        fixPodWithContainers("test-1", "test-2"),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli, cleanup := newFakePodsGetter(t, tc.pod, fixLogsResponse)
			defer cleanup()

			fetcher := logs.NewFetcherForTestingPods(fakeCli, tc.ignoredCnt)

			fix := fixFailedTestResultForPod(tc.pod)

			// when
			gotLogOut, err := fetcher.Logs(fix)

			// then
			require.NoError(t, err)

			assert.Contains(t, gotLogOut, fixLogsResponse)
			assert.Equal(t, fakeCli.podName, tc.pod.Name)
			assert.Equal(t, fakeCli.podNamespace, tc.pod.Namespace)
			assert.NotContains(t, fakeCli.containers, tc.ignoredCnt)
		})
	}
}

func fixFailedTestResultForPod(pod v1.Pod) oct.TestResult {
	return oct.TestResult{
		Namespace: pod.Namespace,
		Status:    oct.TestFailed,
		Executions: []oct.TestExecution{
			{
				ID:       pod.Name,
				PodPhase: v1.PodFailed,
			},
		},
	}

}

func fixPodWithContainers(containersName ...string) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pico",
			Namespace: "bello",
		},
	}

	for _, cn := range containersName {
		pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{
			Name: cn,
		})
	}

	return pod
}

// fakePodGetter implements SIMPLY fake for getting pods and fetching associated logs
// Unfortunately needs to be implemented by us, because client-go fake client
// in current implementation (client-go v1.12) returns zero-value `rest.Request`
// and executing GetLogs(/*..*/).DoRaw() method, throws panic.
type fakePodGetter struct {
	podName      string
	podNamespace string
	containers   []string
	pod          v1.Pod
	url          *url.URL
}

func newFakePodsGetter(t *testing.T, pod v1.Pod, podlogs string) (*fakePodGetter, func()) {
	fakeAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, podlogs)
	}))
	u, err := url.Parse(fakeAPIServer.URL)
	require.NoError(t, err)

	fakeCli := &fakePodGetter{
		pod: pod,
		url: u,
	}

	return fakeCli, fakeAPIServer.Close
}

func (f *fakePodGetter) Pods(namespace string) corev1.PodInterface {
	f.podNamespace = namespace
	return f
}

func (f *fakePodGetter) Get(name string, options metav1.GetOptions) (*v1.Pod, error) {
	f.podName = name
	return &f.pod, nil
}

func (f *fakePodGetter) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	f.containers = append(f.containers, opts.Container)
	return rest.NewRequest(nil, http.MethodGet, f.url, "", rest.ContentConfig{}, rest.Serializers{}, nil, nil, 0)
}

// Defined only to fulfil the `PodInterface` interface
// should not be used by business logic.
func (f *fakePodGetter) Create(*v1.Pod) (*v1.Pod, error) {
	panic("not implemented")
}

func (f *fakePodGetter) Update(*v1.Pod) (*v1.Pod, error) {
	panic("not implemented")
}
func (f *fakePodGetter) UpdateStatus(*v1.Pod) (*v1.Pod, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Delete(name string, options *metav1.DeleteOptions) error {
	panic("not implemented")
}
func (f *fakePodGetter) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not implemented")
}
func (f *fakePodGetter) List(opts metav1.ListOptions) (*v1.PodList, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Pod, err error) {
	panic("not implemented")
}
func (f *fakePodGetter) Bind(binding *v1.Binding) error {
	panic("not implemented")
}
func (f *fakePodGetter) Evict(eviction *v1beta1.Eviction) error {
	panic("not implemented")
}
