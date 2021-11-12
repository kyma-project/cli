package logs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/api/policy/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

func Test(t *testing.T) {
	t.Parallel()
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

			fetcher := NewFetcherForTestingPods(fakeCli, tc.ignoredCnt)

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

func Test_stripANSI(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Got, Wanted string
	}{
		{Got: `[?25l[09:33:25]  Verifying Cypress can run /root/.cache/Cypress/4.12.1/Cypress [started]
[09:33:34]  Verifying Cypress can run /root/.cache/Cypress/4.12.1/Cypress [completed]
[?25hOne login method detected, trying to login using email...`,
			Wanted: `[09:33:25]  Verifying Cypress can run /root/.cache/Cypress/4.12.1/Cypress [started]
[09:33:34]  Verifying Cypress can run /root/.cache/Cypress/4.12.1/Cypress [completed]
One login method detected, trying to login using email...`,
		},
		{
			Got:    "\u001B[4mUnicorn\u001B[0m",
			Wanted: "Unicorn",
		},
		{
			Got:    "\u001B[?25hcode is 1",
			Wanted: "code is 1",
		},
	}
	for _, tc := range testCases {
		s := stripANSI(tc.Got)
		assert.Equal(t, tc.Wanted, s)
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
// in current implementation returns zero-value `rest.Request`
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

func (f *fakePodGetter) Apply(ctx context.Context, pod *coreapplyv1.PodApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Pod, err error) {
	panic("implement me")
}

func (f *fakePodGetter) ApplyStatus(ctx context.Context, pod *coreapplyv1.PodApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Pod, err error) {
	panic("implement me")
}

func (f *fakePodGetter) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	f.podName = name
	return &f.pod, nil
}

func (f *fakePodGetter) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	f.containers = append(f.containers, opts.Container)
	c, err := rest.NewRESTClient(f.url, "", rest.ClientContentConfig{}, flowcontrol.NewFakeAlwaysRateLimiter(), http.DefaultClient)
	if err != nil {
		return nil
	}
	return rest.NewRequest(c)
}

// Defined only to fulfil the `PodInterface` interface
// should not be used by business logic.
func (f *fakePodGetter) Create(ctx context.Context, pod *v1.Pod, opts metav1.CreateOptions) (*v1.Pod, error) {
	panic("not implemented")
}

func (f *fakePodGetter) Update(ctx context.Context, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("not implemented")
}
func (f *fakePodGetter) UpdateStatus(ctx context.Context, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	panic("not implemented")
}
func (f *fakePodGetter) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	panic("not implemented")
}
func (f *fakePodGetter) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Pod, err error) {
	panic("not implemented")
}
func (f *fakePodGetter) Bind(ctx context.Context, binding *v1.Binding, opts metav1.CreateOptions) error {
	panic("not implemented")
}
func (f *fakePodGetter) GetEphemeralContainers(ctx context.Context, podName string, options metav1.GetOptions) (*v1.EphemeralContainer, error) {
	panic("not implemented")
}
func (f *fakePodGetter) UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("not implemented")
}
func (f *fakePodGetter) Evict(ctx context.Context, eviction *v1beta1.Eviction) error {
	panic("not implemented")
}
func (f *fakePodGetter) ProxyGet(scheme, name, port, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (f *fakePodGetter) EvictV1(ctx context.Context, eviction *policyv1.Eviction) error {
	panic("not implemented")
}

func (f *fakePodGetter) EvictV1beta1(ctx context.Context, eviction *policyv1beta1.Eviction) error  {
	panic("not implemented")
}