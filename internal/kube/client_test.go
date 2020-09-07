package kube

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/install/scheme"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

	dynFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIsPodDeployed(t *testing.T) {
	//setup
	c := fakeClientWithNS()
	_, err := c.Static().CoreV1().Pods("ns").Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod1",
		},
	})
	require.NoError(t, err)

	// test finding the pod
	found, err := c.IsPodDeployed("ns", "test-pod1")
	require.NoError(t, err, "Checking if the Pod is deployed and no errors occur.")
	require.True(t, found, "Pod is deployed.")

	// test checking non existing pod
	found, err = c.IsPodDeployed("ns", "non-existing-pod")
	require.NoError(t, err, "Checking if the Pod is deployed and no errors occur.")
	require.False(t, found, "Pod is not deployed")

	// simulate an unexpected error when contacting k8s
	errClient := &fake.Clientset{}
	errClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Error while fetching Pod")
	})

	c.static = errClient
	_, err = c.IsPodDeployed("ns", "test-pod1")
	require.Error(t, err, "Checking if the Pod is deployed and if any errors occur.")
}

func TestIsPodDeployedByLabel(t *testing.T) {
	//setup
	c := fakeClientWithNS()
	_, err := c.Static().CoreV1().Pods("ns").Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-pod1",
			Labels: map[string]string{"team": "huskies"},
		},
	})
	require.NoError(t, err)

	// test finding the pod
	found, err := c.IsPodDeployedByLabel("ns", "team", "huskies")
	require.NoError(t, err, "Checking if the Pod is deployed and no errors occur.")
	require.True(t, found, "Pod is deployed")

	// test checking non existing pod
	found, err = c.IsPodDeployedByLabel("ns", "team", "skydiving-tunas")
	require.NoError(t, err, "Checking if the Pod is deployed and no errors occur.")
	require.False(t, found, "Pod is not deployed")

	// simulate an unexpected error when contacting k8s
	errClient := &fake.Clientset{}
	errClient.Fake.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Error while fetching Pod")
	})

	c.static = errClient
	_, err = c.IsPodDeployedByLabel("ns", "team", "skydiving-tunas")
	require.Error(t, err, "Checking if the Pod is deployed and no errors occur.")

}

func TestWaitPodStatus(t *testing.T) {
	// setup
	c := fakeClientWithNS()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod1",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	_, err := c.Static().CoreV1().Pods("ns").Create(pod)
	require.NoError(t, err)

	// wait for the pod to be running in a separate goroutine
	waitCh := make(chan error)
	go func(ch chan<- error) {
		ch <- c.WaitPodStatus("ns", "test-pod1", corev1.PodRunning)
		close(ch)
	}(waitCh)

	// wait a bit and set the pod to running
	time.Sleep(1 * time.Second)
	pod.Status.Phase = corev1.PodRunning
	_, err = c.Static().CoreV1().Pods("ns").UpdateStatus(pod)
	require.NoError(t, err)

	// we block waiting for the pod to change its state
	require.NoError(t, <-waitCh)
}

func TestWaitPodStatusByLabel(t *testing.T) {
	// setup
	c := fakeClientWithNS()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-pod1",
			Labels: map[string]string{"team": "huskies"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	_, err := c.Static().CoreV1().Pods("ns").Create(pod)
	require.NoError(t, err)

	// wait for the pod to be running in a separate goroutine
	waitCh := make(chan error)
	go func(ch chan<- error) {
		ch <- c.WaitPodStatusByLabel("ns", "team", "huskies", corev1.PodRunning)
		close(ch)
	}(waitCh)

	// wait a bit and set the pod to running
	time.Sleep(1 * time.Second)
	pod.Status.Phase = corev1.PodRunning
	_, err = c.Static().CoreV1().Pods("ns").UpdateStatus(pod)
	require.NoError(t, err)

	// we block waiting for the pod to change its state
	require.NoError(t, <-waitCh)
}

func TestWatchResource(t *testing.T) {
	c := &client{
		cfg: &rest.Config{},
		dynamic: dynamicK8s(
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "fakeAPI/fakeVersion",
					"kind":       "fake",
					"metadata": map[string]interface{}{
						"name": "samus",
					},
					"status": map[string]interface{}{
						"phase": "partying",
					},
				},
			},
		),
	}

	checkFn := func(u *unstructured.Unstructured) (bool, error) {
		status, exists, err := unstructured.NestedString(u.Object, "status", "phase")
		if err != nil {
			return false, err
		}
		return exists && status == "partying", nil
	}

	// non namepsaced
	err := c.WatchResource(schema.GroupVersionResource{Group: "fakeAPI", Version: "fakeVersion", Resource: "fakes"}, "samus", "", checkFn)
	require.NoError(t, err)

	// namespaced
	c.dynamic = dynamicK8s(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "TallonIV",
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "fakeAPI/fakeVersion",
				"kind":       "fake",
				"metadata": map[string]interface{}{
					"name":      "samus",
					"namespace": "TallonIV",
				},
				"status": map[string]interface{}{
					"phase": "partying",
				},
			},
		},
	)
	err = c.WatchResource(schema.GroupVersionResource{Group: "fakeAPI", Version: "fakeVersion", Resource: "fakes"}, "samus", "TallonIV", checkFn)
	require.NoError(t, err)
}

func fakeClientWithNS() *client {
	return &client{
		static: fake.NewSimpleClientset(
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns",
				},
			},
		),
	}
}

func dynamicK8s(objects ...runtime.Object) *dynFake.FakeDynamicClient {
	resSchema, err := scheme.DefaultScheme()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dynFake.NewSimpleDynamicClient(resSchema, objects...)
}
