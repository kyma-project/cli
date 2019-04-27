package helm

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/helm/helm/pkg/tlsutil"
	"github.com/kyma-incubator/kyma-cli/internal/net"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
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
	settings.TLSServerName = settings.TillerHost
	settings.TLSCaCertFile = settings.Home.TLSCaCert()
	settings.TLSCertFile = settings.Home.TLSCert()
	settings.TLSKeyFile = settings.Home.TLSKey()

	options := []helm.Option{helm.Host(settings.TillerHost), helm.ConnectTimeout(settings.TillerConnectionTimeout)}

	if settings.TLSVerify || settings.TLSEnable {
		tlsopts := tlsutil.Options{
			ServerName:         settings.TLSServerName,
			KeyFile:            settings.TLSKeyFile,
			CertFile:           settings.TLSCertFile,
			InsecureSkipVerify: true,
		}
		if settings.TLSVerify {
			tlsopts.CaCertFile = settings.TLSCaCertFile
			tlsopts.InsecureSkipVerify = false
		}
		tlscfg, err := tlsutil.ClientConfig(tlsopts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		options = append(options, helm.WithTLS(tlscfg))
	}
	helmClient := helm.NewClient(options...)

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

func GetHelmHome() (string, error) {
	helmCmd := exec.Command("helm", "home")
	helmHomeRaw, err := helmCmd.CombinedOutput()
	if err != nil {
		return "", nil
	}

	helmHome := strings.Replace(string(helmHomeRaw), "\n", "", -1)
	if _, err := os.Stat(helmHome); os.IsNotExist(err) {
		err = os.MkdirAll(helmHome, 0700)
		if err != nil {
			return "", err
		}
	}
	return helmHome, nil
}
