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
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

//go:embed testdata/kyma_sample.yaml
var kymaTestSample []byte

//go:embed testdata/template_sample.yaml
var moduleTemplateTestSample []byte

func Test_list_modules_without_Kyma(t *testing.T) {

	kyma := &unstructured.Unstructured{}
	moduleTemplate := &unstructured.Unstructured{}

	testCases := []struct {
		name    string
		useKyma bool
		cmd     func(cmd command)
		out     func(t *testing.T, out []byte)
	}{
		{
			"Interactive",
			true,
			func(cmd command) {
				cmd.opts.Output = "go-template-file"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`WARNING: This command is experimental and might change in its final version. Use at your own risk.
operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version		Template		State
manifest-1				kyma.project.io/module/loadtest		stable		0.0.4		default/manifest-1	Ready
`,
					string(out),
				)
			},
		}, {
			"Interactive (All in Cluster)",
			false,
			func(cmd command) {
				cmd.opts.Output = "go-template-file"
				cmd.opts.AllNamespaces = true
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`WARNING: This command is experimental and might change in its final version. Use at your own risk.
operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version		Template		State
manifest-1				kyma.project.io/module/loadtest		stable		0.0.4		default/manifest-1	<no value>
`,
					string(out),
				)
			},
		},
		{
			"Non-Interactive",
			true,
			func(cmd command) {
				cmd.opts.NonInteractive = true
				cmd.opts.Output = "go-template-file"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version		Template		State
manifest-1				kyma.project.io/module/loadtest		stable		0.0.4		default/manifest-1	Ready
`,
					string(out),
				)
			},
		},
		{
			"Non-Interactive & No Headers",
			true,
			func(cmd command) {
				cmd.opts.NonInteractive = true
				cmd.opts.Output = "go-template-file"
				cmd.opts.NoHeaders = true
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`manifest-1	kyma.project.io/module/loadtest		stable	0.0.4	default/manifest-1	Ready
`,
					string(out),
				)
			},
		},
		{
			"YAML",
			true,
			func(cmd command) {
				cmd.opts.NonInteractive = true
				cmd.opts.Output = "yaml"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				outputList := &unstructured.UnstructuredList{}
				a.NoError(yaml.Unmarshal(out, outputList))
				a.Len(outputList.Items, 1)
				a.Equal(moduleTemplate.GetName(), outputList.Items[0].GetName())
			},
		},
		{
			"JSON",
			true,
			func(cmd command) {
				cmd.opts.NonInteractive = true
				cmd.opts.Output = "json"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				outputList := &unstructured.UnstructuredList{}
				a.NoError(json.Unmarshal(out, outputList))
				a.Len(outputList.Items, 1)
				a.Equal(moduleTemplate.GetName(), outputList.Items[0].GetName())
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(
			test.name, func(t *testing.T) {
				a := assert.New(t)
				a.NoError(yaml.Unmarshal(kymaTestSample, kyma))
				a.NoError(yaml.Unmarshal(moduleTemplateTestSample, moduleTemplate))
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
				test.cmd(cmd)

				static := fake.NewSimpleClientset()
				dynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, kyma, moduleTemplate)
				kymaMock.On("Dynamic").Return(dynamic)
				kymaMock.On("Static").Return(static).Once()

				captureStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				var args []string
				if test.useKyma {
					args = append(args, kyma.GetName())
				}
				err := cmd.Run(context.Background(), args)

				a.NoError(w.Close())
				out, _ := io.ReadAll(r)
				os.Stdout = captureStdout
				a.NotEmpty(string(out))

				test.out(t, out)

				a.NoError(err)
			},
		)
	}
}
