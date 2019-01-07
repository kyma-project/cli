package test

import (
	"bytes"
	"fmt"
	"github.com/kyma-incubator/kymactl/internal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Options struct {
}

var testNamespaces = []string{
	"kyma-system",
	"istio-system",
	"knative-serving",
	"kyma-integration",
}

var testReleases = []string{
	"core",
	"monitoring",
	"logging",
	"istio",
	"knative",
	"application-connector",
}

//NewCmd creates a new install command
func NewCmd() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test kyma installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.Run()
		},
	}
	return cmd
}

func (opts *Options) Run() error {
	helmConfig := &environment.EnvSettings{}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return err
	}

	forwarder, err := setupTillerConnection(helmConfig, kubeConfig)
	if err != nil {
		return err
	}
	defer func(){
		if forwarder != nil {
			close(forwarder)
		}
	}()

	helmClient := helm.NewClient(helm.Host(helmConfig.TillerHost), helm.ConnectTimeout(helmConfig.TillerConnectionTimeout))

	for _, ns := range testNamespaces {
		err := opts.cleanHelmTestPods(ns)
		if err != nil {
			return err
		}
	}

	for _, rls := range testReleases {
		err := opts.testRelease(helmClient, rls, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupTillerConnection(settings *environment.EnvSettings, config *rest.Config) (chan<- struct{}, error) {
	if settings.TillerHost != "" {
		return nil, nil
	}

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", "kube-system", "tiller")
	hostIP := strings.TrimLeft(config.Host, "htps://")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	port, err := getAvailablePort()
	if err != nil {
		return nil, err
	}

	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%v:44134", port)}, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, err
	}


	errChan := make(chan error)
	go func() {
		errChan <- forwarder.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return nil, fmt.Errorf("forwarding ports: %v", err)
	case <-forwarder.Ready:
		settings.TillerHost = "127.0.0.1:44134"
		return stopChan, nil
	}
}

func (opts *Options) cleanHelmTestPods(namespace string) error {
	spinner := internal.NewSpinner(
		fmt.Sprintf("Cleaning up helm test pods in namespace %s", namespace),
		fmt.Sprintf("Test pods in namespace %s cleaned", namespace),
	)
	_, err := internal.RunKubectlCmd([]string{"delete", "pod", "-n", namespace, "-l", "helm-chart-test=true"})
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)
	return nil
}

func (opts *Options) testRelease(helm helm.Interface, release string, optional bool) error {
	_, err := helm.ReleaseStatus(release)
	if err != nil {
		if optional {
			return nil
		}
		return err
	}

	rsps, errs := helm.RunReleaseTest(release)
	spinner := internal.NewSpinner(
		fmt.Sprintf("Testing release %s", release),
		fmt.Sprintf("Release %s tests success", release),
	)

	for {
		select {
		case rsp := <-rsps:
			switch rsp.Status {
			case hapi_release.TestRun_SUCCESS:
				internal.StopSpinner(spinner)
				return nil
			case hapi_release.TestRun_FAILURE:
				return errors.Errorf("FAILED: %s", rsp.Msg)
			default:
				continue
			}
		case err := <-errs:
			return err
		}
	}
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}
