package module

import (
	"context"
	_ "embed"
	"io"
	"os"
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

//go:embed kyma_sample.yaml
var kymaTestSample []byte

//go:embed template_sample.yaml
var moduleTemplateTestSample []byte

func Test_list_modules_without_Kyma(t *testing.T) {
	a := assert.New(t)
	kymaMock := &mocks.KymaKube{}
	cmd := command{
		Command: cli.Command{
			K8s: kymaMock,
		},
		opts: &Options{Options: cli.NewOptions(),
			Timeout:   1 * time.Minute,
			Namespace: metav1.NamespaceDefault,
		},
	}
	cmd.opts.NonInteractive = true

	kyma := &unstructured.Unstructured{}
	a.NoError(yaml.Unmarshal(kymaTestSample, kyma))
	moduleTemplate := &unstructured.Unstructured{}
	a.NoError(yaml.Unmarshal(moduleTemplateTestSample, moduleTemplate))
	static := fake.NewSimpleClientset()
	dynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, kyma, moduleTemplate)
	kymaMock.On("Dynamic").Return(dynamic)
	kymaMock.On("Static").Return(static).Once()

	captureStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := cmd.Run(context.Background(), []string{kyma.GetName()})

	a.NoError(w.Close())
	out, _ := io.ReadAll(r)
	os.Stdout = captureStdout
	a.NotEmpty(string(out))

	outputList := &unstructured.UnstructuredList{}
	a.NoError(yaml.Unmarshal(out, outputList))

	a.Len(outputList.Items, 1)
	a.Equal(outputList.Items[0].GetName(), moduleTemplate.GetName())

	a.NoError(err)
}
