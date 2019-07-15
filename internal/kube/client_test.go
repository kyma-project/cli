package kube

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"

	k8stesting "k8s.io/client-go/testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestIsPodDeployed(t *testing.T) {
	//setup
	c := fakeClientWithNS()
	c.Static().CoreV1().Pods("ns").Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod1",
		},
	})

	// test finding the pod
	found, err := c.IsPodDeployed("ns", "test-pod1")
	require.NoError(t, err, "Checking if a pod is deployed error not as expected.")
	require.True(t, found, "Checking if a pod is deployed should be true")

	// test checking non existing pod
	found, err = c.IsPodDeployed("ns", "non-existing-pod")
	require.NoError(t, err, "Checking if a pod is deployed error not as expected.")
	require.False(t, found, "Checking if a pod is deployed should be false")

	// simulate an unexpected error when contacting k8s
	errClient := &fake.Clientset{}
	errClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Error while fetching Pod")
	})

	c.static = errClient
	_, err = c.IsPodDeployed("ns", "test-pod1")
	require.Error(t, err, "Checking if a pod is deployed error not as expected.")
}

func TestIsPodDeployedByLabel(t *testing.T) {
	//setup
	c := fakeClientWithNS()
	c.Static().CoreV1().Pods("ns").Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-pod1",
			Labels: map[string]string{"team": "huskies"},
		},
	})

	// test finding the pod
	found, err := c.IsPodDeployedByLabel("ns", "team", "huskies")
	require.NoError(t, err, "Checking if a pod is deployed error not as expected.")
	require.True(t, found, "Checking if a pod is deployed should be true")

	// test checking non existing pod
	found, err = c.IsPodDeployedByLabel("ns", "team", "skydiving-tunas")
	require.NoError(t, err, "Checking if a pod is deployed error not as expected.")
	require.False(t, found, "Checking if a pod is deployed should be false")

	// simulate an unexpected error when contacting k8s
	errClient := &fake.Clientset{}
	errClient.Fake.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Error while fetching Pod")
	})

	c.static = errClient
	_, err = c.IsPodDeployedByLabel("ns", "team", "skydiving-tunas")
	require.Error(t, err, "Checking if a pod is deployed error not as expected.")

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
	c.Static().CoreV1().Pods("ns").Create(pod)

	// wait for the pod to be running in a separate goroutine
	waitCh := make(chan error)
	go func(ch chan<- error) {
		ch <- c.WaitPodStatus("ns", "test-pod1", corev1.PodRunning)
		close(ch)
	}(waitCh)

	// wait a bit and set the pod to running
	time.Sleep(1 * time.Second)
	pod.Status.Phase = corev1.PodRunning
	c.Static().CoreV1().Pods("ns").UpdateStatus(pod)

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
	c.Static().CoreV1().Pods("ns").Create(pod)

	// wait for the pod to be running in a separate goroutine
	waitCh := make(chan error)
	go func(ch chan<- error) {
		ch <- c.WaitPodStatusByLabel("ns", "team", "huskies", corev1.PodRunning)
		close(ch)
	}(waitCh)

	// wait a bit and set the pod to running
	time.Sleep(1 * time.Second)
	pod.Status.Phase = corev1.PodRunning
	c.Static().CoreV1().Pods("ns").UpdateStatus(pod)

	// we block waiting for the pod to change its state
	require.NoError(t, <-waitCh)
}

func fakeClientWithNS() *client {
	c := &client{
		static: fake.NewSimpleClientset(),
	}

	c.Static().CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns",
		},
	})

	return c
}
