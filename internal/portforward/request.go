package portforward

import (
	"net/http"

	"k8s.io/client-go/rest"
)

func DoRequest(config *rest.Config, podName, podNamespace, podPort string, request *http.Request) (*http.Response, error) {
	conn, err := NewDialFor(config, podName, podNamespace)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	transport := NewPortforwardTransport(conn, podPort)
	client := &http.Client{Transport: transport}

	return client.Do(request)
}
