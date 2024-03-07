package portforward

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	client_portforward "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func NewDialFor(config *rest.Config, podName, podNamespace string) (httpstream.Connection, error) {
	portforwardAddr := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward",
		config.Host,
		podNamespace,
		podName,
	)

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	portForwardURL, err := url.Parse(portforwardAddr)
	if err != nil {
		return nil, err
	}

	cl := &http.Client{Transport: transport}
	dialer := spdy.NewDialer(upgrader, cl, "POST", portForwardURL)

	conn, _, err := dialer.Dial(client_portforward.PortForwardProtocolV1Name)
	return conn, err
}
