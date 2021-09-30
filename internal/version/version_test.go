package version

import (
	"github.com/kyma-project/cli/internal/kube/mocks"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	t.Run("Create NoVersion", func(t *testing.T) {
		version := NewNoVersion()
		assert.True(t, version.None())
	})

	t.Run("Create from Kyma 2 version", func(t *testing.T) {
		version, err := NewKymaVersion("2.01")
		assert.NoError(t, err)
		assert.True(t, version.IsKyma2())
	})

	t.Run("Create from Kyma 1 version", func(t *testing.T) {
		version, err := NewKymaVersion("1.024")
		assert.NoError(t, err)
		assert.True(t, version.IsKyma1())
	})

	t.Run("Create from PR version", func(t *testing.T) {
		version, err := NewKymaVersion("PR-123")
		assert.NoError(t, err)
		assert.False(t, version.IsReleasedVersion())
	})

	t.Run("Returns same version as string", func(t *testing.T) {
		version := "123"
		kymaVersion, err := NewKymaVersion(version)
		assert.NoError(t, err)
		assert.Equal(t, version, kymaVersion.String())
	})
}

func TestCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("Check same version", func(t *testing.T) {
		kymaVersion, err := NewKymaVersion("2.0.0")
		assert.NoError(t, err)
		upgKymaVersion, err := NewKymaVersion("2.0.0")
		assert.NoError(t, err)
		res := kymaVersion.IsCompatibleWith(upgKymaVersion)
		assert.Equal(t, UpgradeEqualVersion, res)
	})
	t.Run("Check undetermined version", func(t *testing.T) {
		kymaVersion, err := NewKymaVersion("abcde")
		assert.NoError(t, err)
		upgKymaVersion, err := NewKymaVersion("2.0.0")
		assert.NoError(t, err)
		res := kymaVersion.IsCompatibleWith(upgKymaVersion)
		assert.Equal(t, UpgradeUndetermined, res)
	})
	t.Run("Check upgrade possible", func(t *testing.T) {
		kymaVersion, err := NewKymaVersion("2.0.0")
		assert.NoError(t, err)
		upgKymaVersion, err := NewKymaVersion("2.1.0")
		assert.NoError(t, err)
		res := kymaVersion.IsCompatibleWith(upgKymaVersion)
		assert.Equal(t, UpgradePossible, res)
	})
}
func TestGetCurrentVersion(t *testing.T) {
	kymaMock := &mocks.KymaKube{}
	t.Run("Check for kyma 2 version", func(t *testing.T) {
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
		kymaMock.On("Static").Return(mockDep).Once()
		kymaMock.On("Static").Return(mockDep).Once()
		res, err := GetCurrentKymaVersion(kymaMock)
		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", res.String())
	})

	t.Run("Check for kyma 1 version", func(t *testing.T) {
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
		res, err := GetCurrentKymaVersion(kymaMock)
		assert.NoError(t, err)
		assert.Equal(t, "1.24.6", res.String())
	})
	t.Run("No Kyma installed", func(t *testing.T) {
		mockClientSet := fake.NewSimpleClientset(		)
		kymaMock.On("Static").Return(mockClientSet).Once()
		kymaMock.On("Static").Return(mockClientSet).Once()
		res, err := GetCurrentKymaVersion(kymaMock)
		assert.NoError(t, err)
		assert.Equal(t, "N/A", res.String())
	})

	t.Run("Non semver", func(t *testing.T) {
		var l = make(map[string]string)
		l["reconciler.kyma-project.io/managed-by"] = "reconciler"
		l["reconciler.kyma-project.io/origin-version"] = "2.a"
		mockDep := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kyma-system",
					Labels:    l,
				},
			},
		)
		kymaMock.On("Static").Return(mockDep).Once()
		kymaMock.On("Static").Return(mockDep).Once()
		_, err := GetCurrentKymaVersion(kymaMock)
		t.Logf("err :%v", err)
		assert.Error(t, err,"Version is not a semver")
	})

}