package integration

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"path"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"time"

	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"testing"
)

var testenv env.Environment

const (
	controllerDeploymentName = "lifecycle-manager-controller-manager"
	kcpNamespace             = "kcp-system"
	kymaNamespace            = "kyma-system"
	kymaName                 = "default-kyma"
)

func TestMain(m *testing.M) {
	cfg, err := envconf.NewFromFlags()
	if err != nil {
	}
	cfg = cfg.WithKubeconfigFile(conf.ResolveKubeConfigFile())
	testenv = env.NewWithConfig(cfg)
	os.Exit(testenv.Run(m))
}

func TestAlphaCommands(t *testing.T) {
	featDeploy := features.New("alpha deploy").
		WithStep("command exec", 1, alphaDeployStep).
		Assess("controller deployment exists", deploymentExists(kcpNamespace, controllerDeploymentName)).
		Assess("controller deployment is available", deploymentAvailable(kcpNamespace, controllerDeploymentName)).
		Assess("kyma is ready", kymaReady(kymaNamespace, kymaName)).
		Feature()

	testenv.Test(t, featDeploy)
}

func deploymentExists(namespace, name string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()
		client, err := cfg.NewClient()
		if err != nil {
			t.Fatal(err)
		}
		dep := controllerManagerDeployment(namespace, name)
		err = wait.For(
			conditions.New(
				client.Resources(),
			).ResourcesFound(&appsv1.DeploymentList{Items: []appsv1.Deployment{dep}}),
		)
		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

func kymaReady(namespace string, name string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()
		resourcesFromConfig, err := resources.New(cfg.Client().RESTConfig())
		if err != nil {
			t.Fatal(err)
		}
		if err := v1beta1.AddToScheme(resourcesFromConfig.GetScheme()); err != nil {
			t.Fatal(err)
		}

		var kyma v1beta1.Kyma
		if err := wait.For(func() (bool, error) {
			if err := resourcesFromConfig.Get(ctx, name, namespace, &kyma); err != nil {
				t.Fatal(err)
			}
			return kyma.Status.State == v1beta1.StateReady, nil
		}); err != nil {
			t.Fatal(err)
		}
		logKymaStatus(ctx, t, resourcesFromConfig, kyma)

		return ctx
	}
}

func deploymentAvailable(namespace, name string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()
		client, err := cfg.NewClient()
		if err != nil {
			t.Fatal(err)
		}
		deployment := controllerManagerDeployment(namespace, name)
		err = wait.For(
			conditions.New(client.Resources()).DeploymentConditionMatch(
				deployment.DeepCopy(),
				appsv1.DeploymentAvailable, corev1.ConditionTrue,
			),
			wait.WithTimeout(time.Minute*3),
		)
		if err != nil {
			t.Fatal(err)
		}

		pods := corev1.PodList{}
		_ = client.Resources(namespace).List(ctx, &pods, func(options *metav1.ListOptions) {
			sel, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
			if err != nil {
				t.Fatal(err)
			}
			options.LabelSelector = sel.String()
		})

		for _, pod := range pods.Items {
			if marshal, err := yaml.Marshal(&pod.Status); err == nil {
				t.Logf("Pod Status Name %s/%s\n%s", pod.Namespace, pod.Name, marshal)
			}
		}
		logDeployStatus(ctx, t, client, deployment)

		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}

func logKymaStatus(ctx context.Context, t *testing.T, r *resources.Resources, kyma v1beta1.Kyma) {
	t.Helper()
	errCheckCtx, cancelErrCheck := context.WithTimeout(ctx, 5*time.Second)
	defer cancelErrCheck()
	if err := r.Get(errCheckCtx, kyma.Name, kyma.Namespace, &kyma); err != nil {
		t.Error(err)
	}
	if marshal, err := yaml.Marshal(&kyma.Status); err == nil {
		t.Logf("%s", marshal)
	}
}

func logDeployStatus(ctx context.Context, t *testing.T, client klient.Client, dep appsv1.Deployment) {
	t.Helper()
	errCheckCtx, cancelErrCheck := context.WithTimeout(ctx, 5*time.Second)
	defer cancelErrCheck()
	if err := client.Resources().Get(errCheckCtx, dep.Name, dep.Namespace, &dep); err != nil {
		t.Error(err)
	}
	if marshal, err := yaml.Marshal(&dep.Status); err == nil {
		t.Logf("%s", marshal)
	}
}

func controllerManagerDeployment(namespace string, name string) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace,
			Labels: map[string]string{"app.kubernetes.io/component": "lifecycle-manager.kyma-project.io"},
		},
	}
}

var alphaDeployStep = func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
	err := execKymaAlphaCmd("deploy")
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func execKymaAlphaCmd(args ...string) error {
	curr, err := os.Getwd()
	if err != nil {
		return err
	}
	binPath := fmt.Sprintf("%s%s", path.Join(path.Dir(path.Dir(curr)), "bin"), "/kyma-darwin")
	args = append([]string{"alpha", "--ci"}, args...)
	fmt.Printf("Executing: %s %s\n", binPath, strings.Join(args, " "))
	fmt.Println("---")
	cmd := exec.Command(binPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdoutReader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	var n int
	for err == nil {
		n, err = stdoutReader.Read(buffer)
		if n > 0 {
			fmt.Printf(string(buffer[0:n]))
		}
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	fmt.Println("---")
	return nil
}
