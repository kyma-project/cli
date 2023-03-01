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
	"github.com/kyma-project/lifecycle-manager/api"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	fake2 "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//go:embed testdata/kyma_sample.yaml
var kymaTestSample []byte

//go:embed testdata/template_sample.yaml
var moduleTemplateTestSample []byte

func Test_list_modules_without_Kyma(t *testing.T) {

	kyma := &v1beta1.Kyma{}
	moduleTemplate := &v1beta1.ModuleTemplate{}

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
				cmd.opts.Output = "tabwriter"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`WARNING: This command is experimental and might change in its final version. Use at your own risk.
Template		operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version		State
kcp-system/manifest-1	manifest-1				kyma.project.io/module/loadtest		stable		0.0.4		Ready
`,
					string(out),
				)
			},
		}, {
			"Interactive (Namespace specific)",
			false,
			func(cmd command) {
				cmd.opts.Output = "tabwriter"
				cmd.opts.Namespace = "kcp-system"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`WARNING: This command is experimental and might change in its final version. Use at your own risk.
Template		operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version
kcp-system/manifest-1	manifest-1				kyma.project.io/module/loadtest		stable		0.0.4
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
				cmd.opts.Output = "tabwriter"
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`Template		operator.kyma-project.io/module-name	Domain Name (FQDN)			Channel		Version		State
kcp-system/manifest-1	manifest-1				kyma.project.io/module/loadtest		stable		0.0.4		Ready
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
				cmd.opts.Output = "tabwriter"
				cmd.opts.NoHeaders = true
			},
			func(t *testing.T, out []byte) {
				a := assert.New(t)
				a.Equal(
					`kcp-system/manifest-1	manifest-1	kyma.project.io/module/loadtest		stable	0.0.4	Ready
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
				outputList := &v1beta1.ModuleTemplateList{}
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
				outputList := &v1beta1.ModuleTemplateList{}
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
						Options: cli.NewOptions(),
						K8s:     kymaMock,
					},
					opts: &Options{Options: cli.NewOptions(),
						Timeout:   1 * time.Minute,
						Namespace: metav1.NamespaceAll,
					},
				}
				test.cmd(cmd)

				a.NoError(api.AddToScheme(scheme.Scheme))

				ctrlClient := fake2.NewClientBuilder().WithScheme(scheme.Scheme).
					WithObjects(kyma).
					WithLists(&v1beta1.ModuleTemplateList{Items: []v1beta1.ModuleTemplate{*moduleTemplate}}).Build()
				static := fake.NewSimpleClientset()
				dynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, kyma, moduleTemplate)
				kymaMock.On("Dynamic").Return(dynamic)
				kymaMock.On("Static").Return(static).Once()
				kymaMock.On("Ctrl").Return(ctrlClient)

				captureStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				var args []string
				if test.useKyma {
					cmd.opts.Namespace = kyma.GetNamespace()
					cmd.opts.KymaName = kyma.GetName()
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
