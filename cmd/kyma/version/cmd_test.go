package version

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/kyma-project/cli/internal/version"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestKyma2Version(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
		},
	}
	cmd.Factory.NonInteractive = true

	var l = make(map[string]string)
	l["reconciler.kyma-project.io/managed-by"] = "reconciler"
	l["reconciler.kyma-project.io/origin-version"] = "2.0.0"

	mockDep := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "kyma-system",
				Labels:    l,
			},
		},
	)
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the version
	kymaMock.On("Static").Return(mockDep).Once()
	kymaMock.On("Static").Return(mockDep).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "2.0.0", ver.String())
}

func TestKyma1Version(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
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
				Name:      "kyma-installer",
				Namespace: "kyma-installer",
				Labels:    l,
			}, Spec: coreV1.PodSpec{
				Containers: []coreV1.Container{con},
			},
		},
	)
	kymaMock.On("Static").Return(mockPod).Once()
	kymaMock.On("Static").Return(mockPod).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "1.24.6", ver.String())
}

func TestPRVersion(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
		},
	}
	cmd.Factory.NonInteractive = true

	var l = make(map[string]string)
	l["reconciler.kyma-project.io/managed-by"] = "reconciler"
	l["reconciler.kyma-project.io/origin-version"] = "PR-12345"

	mockDep := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "kyma-system",
				Labels:    l,
			},
		},
	)
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the version
	kymaMock.On("Static").Return(mockDep).Once()
	kymaMock.On("Static").Return(mockDep).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "PR-12345", ver.String())
}

func TestSHAVersion(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
		},
	}
	cmd.Factory.NonInteractive = true

	var l = make(map[string]string)
	l["reconciler.kyma-project.io/managed-by"] = "reconciler"
	l["reconciler.kyma-project.io/origin-version"] = "main-abcde"

	mockDep := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "kyma-system",
				Labels:    l,
			},
		},
	)
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the version
	kymaMock.On("Static").Return(mockDep).Once()
	kymaMock.On("Static").Return(mockDep).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "main-abcde", ver.String())
}

func TestBranchVersion(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
		},
	}
	cmd.Factory.NonInteractive = true

	var l = make(map[string]string)
	l["reconciler.kyma-project.io/managed-by"] = "reconciler"
	l["reconciler.kyma-project.io/origin-version"] = "my-branch"

	mockDep := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "kyma-system",
				Labels:    l,
			},
		},
	)
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the version
	kymaMock.On("Static").Return(mockDep).Once()
	kymaMock.On("Static").Return(mockDep).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "my-branch", ver.String())
}

func TestNoKymaVersion(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			Options: cli.NewOptions(),
			K8s:     kymaMock,
		},
	}
	cmd.Factory.NonInteractive = true

	mockClientset := fake.NewSimpleClientset()
	// the kubeclient needs to be faked twice since 1. it checks the kymaVersion and 2. it checks the version
	kymaMock.On("Static").Return(mockClientset).Once()
	kymaMock.On("Static").Return(mockClientset).Once()

	ver, err := version.GetCurrentKymaVersion(cmd.K8s)
	require.NoError(t, err)
	require.Equal(t, "N/A", ver.String())
}