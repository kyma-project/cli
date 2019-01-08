package helm

import (
	"bytes"
	"fmt"
	"github.com/kyma-incubator/kymactl/internal/net"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"net/http"
	"net/url"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Client struct {
	helm.Interface
	forwarder chan<- struct{}
}

func New(settings *environment.EnvSettings, config *rest.Config) (*Client, error) {
	forwarder, err := setupTillerConnection(settings, config)
	if err != nil {
		return nil, err
	}
	helmClient := helm.NewClient(helm.Host(settings.TillerHost), helm.ConnectTimeout(settings.TillerConnectionTimeout))
	return &Client{
		Interface: helmClient,
		forwarder: forwarder,
	}, nil
}

func setupTillerConnection(settings *environment.EnvSettings, config *rest.Config) (chan<- struct{}, error) {
	if settings.TillerHost != "" {
		return nil, nil
	}

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podList, err := clientset.CoreV1().Pods("kube-system").List(meta_v1.ListOptions{LabelSelector: "app=helm"})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, errors.New("tiller not installed")
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", "kube-system", podList.Items[0].Name)
	serverURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}
	serverURL.Path = path

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	port, err := net.GetAvailablePort()
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
		return nil, errors.Errorf("forwarding ports: %v", err)
	case <-forwarder.Ready:
		settings.TillerHost = fmt.Sprintf("127.0.0.1:%v", port)
		return stopChan, nil
	}
}

func (c *Client) Close() {
	if c.forwarder != nil {
		close(c.forwarder)
	}
}
