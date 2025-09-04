package portforward

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
)

// portforwardTransport forwards requests sent by the client to the port-forwarded pod from the cluster
type portforwardTransport struct {
	remoteConn httpstream.Connection
	remotePort string
}

// NewPortforwardTransport creates http.RoundTripper implementation for given portforward connection and port
func NewPortforwardTransport(remoteConn httpstream.Connection, remotePort string) http.RoundTripper {
	return &portforwardTransport{
		remoteConn: remoteConn,
		remotePort: remotePort,
	}
}

func (pft *portforwardTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return pft.dialRemote(req)
}

// dialRemote is forwarding any incoming request to the stream
// the logic is mostly based on the k8s.io/client-go/tools/portforward package
// https://github.com/kubernetes/client-go/blob/271d034e86108101a804541843d50abe3fea06ae/tools/portforward/portforward.go#L335
func (pft *portforwardTransport) dialRemote(req *http.Request) (*http.Response, error) {
	forwardID := rand.Int()

	// create error stream
	errorStream, err := pft.createErrorStream(forwardID)
	if err != nil {
		return nil, fmt.Errorf("error creating error stream for port %s: %v", pft.remotePort, err)
	}
	// close stream to inform remote server that we are not going to send any data,
	// and that we are ready to receive the errors
	errorStream.Close()
	defer pft.remoteConn.RemoveStreams(errorStream)

	errorChan := make(chan error)
	go handleErrorStream(errorStream, errorChan)

	// create data stream
	dataStream, err := pft.createDataStream(forwardID)
	if err != nil {
		return nil, fmt.Errorf("error creating data stream for port %s: %v", pft.remotePort, err)
	}
	defer dataStream.Close()
	defer pft.remoteConn.RemoveStreams(dataStream)

	resp, err := dialToStream(dataStream, req)
	if err != nil {
		return nil, err
	}

	// always expect something on errorChan (it may be nil)
	err = <-errorChan

	return resp, err
}

func (pft *portforwardTransport) createErrorStream(forwardID int) (httpstream.Stream, error) {
	return createStream(pft.remoteConn, pft.remotePort, v1.StreamTypeError, forwardID)
}

func (pft *portforwardTransport) createDataStream(forwardID int) (httpstream.Stream, error) {
	return createStream(pft.remoteConn, pft.remotePort, v1.StreamTypeData, forwardID)
}

func createStream(remoteConn httpstream.Connection, remotePort, streamType string, forwardID int) (httpstream.Stream, error) {
	headers := http.Header{}
	headers.Set(v1.StreamType, streamType)
	headers.Set(v1.PortHeader, remotePort)
	headers.Set(v1.PortForwardRequestIDHeader, strconv.Itoa(forwardID))

	return remoteConn.CreateStream(headers)
}

func dialToStream(dataStream httpstream.Stream, req *http.Request) (*http.Response, error) {
	if err := req.Write(dataStream); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		return nil, fmt.Errorf("error writing to remote stream: %v", err)
	}

	// close stream to inform remote server that there is nothing more to receive
	dataStream.Close()

	resp, err := http.ReadResponse(bufio.NewReader(dataStream), req)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		return nil, fmt.Errorf("error reading from remote stream: %v", err)
	}

	return resp, nil
}

func handleErrorStream(errorStream httpstream.Stream, errorChan chan error) {
	message, err := io.ReadAll(errorStream)
	switch {
	case err != nil:
		errorChan <- fmt.Errorf("error reading from error stream: %v", err)
	case len(message) > 0:
		errorChan <- fmt.Errorf("an error occurred while forwarding: %v", string(message))
	}
	close(errorChan)
}
