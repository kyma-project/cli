package deploy

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"testing"
)

func Test_upgkyma2tokyma2(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:         kymaMock,
		},
		opts: &Options{
			Source:         "main",
		},
	}
	cmd.Factory.NonInteractive = true

	var l = make(map[string]string)
	l["reconciler.kyma-project.io/managed-by"] = "reconciler"
	l["reconciler.kyma-project.io/origin-version"] = "2.0.0"

	mockDep := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:                       "foo",
				Namespace:                  "kyma-system",
				Labels:                     l,
			},
		},
	)
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the versiob
	kymaMock.On("Static").Return(mockDep).Once()
	kymaMock.On("Static").Return(mockDep).Once()

	captureStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := cmd.isCompatibleVersion()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout =captureStdout
	expectedOutput := "? A kyma v2 installation (2.0.0) was found. Do you want to proceed with the upgrade? Type [y/N]:"
	require.Contains(t, string(out), expectedOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Upgrade stopped by user")
}

func Test_upgkyma1tokyma2(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:         kymaMock,
		},
		opts: &Options{
			Source:         "2.0.0",
		},
	}
	cmd.Factory.NonInteractive = false
	var l = make(map[string]string)
	l["name"] = "kyma-installer"

	con := coreV1.Container{}
	con.Image = "foo:1.24.6"

	mockPod := fake.NewSimpleClientset(
		&coreV1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:                       "kyma-installer",
				Namespace:                  "kyma-installer",
				Labels: l,
			},Spec: coreV1.PodSpec{
				Containers:                    []coreV1.Container{con},

			},
		},
	)
	kymaMock.On("Static").Return(mockPod).Once()
	kymaMock.On("Static").Return(mockPod).Once()

	captureStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := cmd.isCompatibleVersion()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = captureStdout
	expectedOutput := "? A kyma v1 installation (1.24.6) was found. Do you want to proceed with the upgrade (2.0.0)? Type [y/N]:"
	require.Contains(t, string(out), expectedOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Upgrade stopped by user")
}