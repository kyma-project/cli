package system

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyma-incubator/hydroform/install/scheme"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateArgs(t *testing.T) {
	t.Parallel()
	c := NewCmd(NewOptions(nil))

	// no args
	err := c.ValidateArgs(nil)
	require.Error(t, err, "Validate args should return an error if no args are given.")

	// 1 arg
	err = c.ValidateArgs([]string{"sys-name"})
	require.NoError(t, err, "Validate args should not return error with 1 argument.")

	// too many args
	err = c.ValidateArgs([]string{"sys-name", "arg2"})
	require.Error(t, err, "Validate args should return an error if too many args are given.")
}

func TestSteps(t *testing.T) {
	t.Parallel()
	c := command{
		Command: cli.Command{
			Options: cli.NewOptions(nil),
		},
		opts: NewOptions(nil),
	}

	// with default output steps are created and print to StdOut
	c.newStep("test msg")
	require.NotNil(t, c.CurrentStep, "On default output steps should be used")
	// success and fail should not crash
	c.successStep("")
	c.failStep()

	// with YAML output steps are omitted
	c.opts.OutputFormat = "yaml"
	c.CurrentStep = nil
	c.newStep("test msg")
	require.Nil(t, c.CurrentStep, "On yaml output steps should be omitted")
	// success and fail should not panic
	c.successStep("")
	c.failStep()
}

func TestCreateSystem(t *testing.T) {
	t.Parallel()
	k8s := &mocks.KymaKube{}
	// mock watch because we can't mock the operator doing changes to the resource in real time
	k8s.On("WatchResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// happy path
	dyn := dynamicK8s()
	k8s.On("Dynamic").Return(dyn).Times(3)

	s, err := createSystem("sys", false, k8s)
	require.NoError(t, err, "Happy path should have no errors.")
	name := fieldOrFail(t, s, "metadata", "name")
	require.Equal(t, "sys", name, "Returned system name not as expected.")

	// system already exists
	dyn = dynamicK8s(
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "applicationconnector.kyma-project.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "sys",
				},
			},
		},
	)
	k8s.On("Dynamic").Return(dyn).Times(3)

	_, err = createSystem("sys", false, k8s)
	require.Error(t, err, "If system already exists and no update flag is passed, an error is expected.")

	// update system
	k8s.On("Dynamic").Return(dyn).Times(3)

	s, err = createSystem("sys", true, k8s)
	require.NoError(t, err, "If system already exists and update flag is passed, no error is expected.")
	name = fieldOrFail(t, s, "metadata", "name")
	require.Equal(t, "sys", name, "System name should be the same after an update.")

}

func TestBindNamespace(t *testing.T) {
	t.Parallel()
	k8s := &mocks.KymaKube{}

	dyn := dynamicK8s(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns",
			},
		},
	)

	// happy path
	k8s.On("Dynamic").Return(dyn)
	err := bindNamespace("mapping", "ns", k8s)
	require.NoError(t, err, "Happy path should not have errors.")

	// mapping already exists
	err = bindNamespace("mapping", "ns", k8s)
	require.NoError(t, err, "Overwritting a mapping should not have errors.")
}

func TestCreateToken(t *testing.T) {
	t.Parallel()
	k8s := &mocks.KymaKube{}
	// mock watch because we can't mock the operator doing changes to the resource in real time
	k8s.On("WatchResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	dyn := dynamicK8s(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns",
			},
		},
	)
	k8s.On("Dynamic").Return(dyn)

	// happy path
	token, err := createToken("tk", "ns", k8s)
	require.NoError(t, err, "Happy path should not have errors.")
	require.NotNil(t, token, "Happy path token should not be nil.")

	// token already exists
	token, err = createToken("tk", "ns", k8s)
	require.NoError(t, err, "Refreshing a token should not have errors.")
	require.NotNil(t, token, "Refreshed token should not be nil.")
}

/* --- Helpers --- */

func dynamicK8s(objects ...runtime.Object) *fake.FakeDynamicClient {
	resSchema, err := scheme.DefaultScheme()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return fake.NewSimpleDynamicClient(resSchema, objects...)
}

func fieldOrFail(t *testing.T, u *unstructured.Unstructured, fields ...string) string {
	f, exists, err := unstructured.NestedString(u.Object, fields...)
	require.NoError(t, err)
	require.True(t, exists)
	return f
}
